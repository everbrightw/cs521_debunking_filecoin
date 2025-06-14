//go:build !nofuse && !openbsd && !netbsd && !plan9
// +build !nofuse,!openbsd,!netbsd,!plan9

package readonly

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	gopath "path"
	"strings"
	"sync"
	"testing"

	"bazil.org/fuse"

	core "github.com/ipfs/kubo/core"
	coreapi "github.com/ipfs/kubo/core/coreapi"
	coremock "github.com/ipfs/kubo/core/mock"

	fstest "bazil.org/fuse/fs/fstestutil"
	chunker "github.com/ipfs/kubo/chunk"
	"github.com/ipfs/boxo/files"
	dag "github.com/ipfs/boxo/ipld/merkledag"
	importer "github.com/ipfs/boxo/ipld/unixfs/importer"
	uio "github.com/ipfs/boxo/ipld/unixfs/io"
	"github.com/ipfs/boxo/path"
	ipld "github.com/ipfs/go-ipld-format"
	"github.com/ipfs/go-test/random"
	ci "github.com/libp2p/go-libp2p-testing/ci"
)

func maybeSkipFuseTests(t *testing.T) {
	if ci.NoFuse() {
		t.Skip("Skipping FUSE tests")
	}
}

func randObj(t *testing.T, nd *core.IpfsNode, size int64) (ipld.Node, []byte) {
	buf := make([]byte, size)
	_, err := io.ReadFull(random.NewRand(), buf)
	if err != nil {
		t.Fatal(err)
	}
	read := bytes.NewReader(buf)
	obj, err := importer.BuildTrickleDagFromReader(nd.DAG, chunker.DefaultSplitter(read))
	if err != nil {
		t.Fatal(err)
	}

	return obj, buf
}

func setupIpfsTest(t *testing.T, node *core.IpfsNode) (*core.IpfsNode, *fstest.Mount) {
	t.Helper()
	maybeSkipFuseTests(t)

	var err error
	if node == nil {
		node, err = coremock.NewMockNode()
		if err != nil {
			t.Fatal(err)
		}
	}

	fs := NewFileSystem(node)
	mnt, err := fstest.MountedT(t, fs, nil)
	if err == fuse.ErrOSXFUSENotFound {
		t.Skip(err)
	}
	if err != nil {
		t.Fatalf("error mounting temporary directory: %v", err)
	}

	return node, mnt
}

// Test writing an object and reading it back through fuse.
func TestIpfsBasicRead(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	nd, mnt := setupIpfsTest(t, nil)
	defer mnt.Close()

	fi, data := randObj(t, nd, 10000)
	k := fi.Cid()
	fname := gopath.Join(mnt.Dir, k.String())
	rbuf, err := os.ReadFile(fname)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(rbuf, data) {
		t.Fatal("Incorrect Read!")
	}
}

func getPaths(t *testing.T, ipfs *core.IpfsNode, name string, n *dag.ProtoNode) []string {
	if len(n.Links()) == 0 {
		return []string{name}
	}
	var out []string
	for _, lnk := range n.Links() {
		child, err := lnk.GetNode(ipfs.Context(), ipfs.DAG)
		if err != nil {
			t.Fatal(err)
		}

		childpb, ok := child.(*dag.ProtoNode)
		if !ok {
			t.Fatal(dag.ErrNotProtobuf)
		}

		sub := getPaths(t, ipfs, gopath.Join(name, lnk.Name), childpb)
		out = append(out, sub...)
	}
	return out
}

// Perform a large number of concurrent reads to stress the system.
func TestIpfsStressRead(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	nd, mnt := setupIpfsTest(t, nil)
	defer mnt.Close()

	api, err := coreapi.NewCoreAPI(nd)
	if err != nil {
		t.Fatal(err)
	}

	var nodes []ipld.Node
	var paths []string

	nobj := 50
	ndiriter := 50

	// Make a bunch of objects
	for i := 0; i < nobj; i++ {
		fi, _ := randObj(t, nd, rand.Int63n(50000))
		nodes = append(nodes, fi)
		paths = append(paths, fi.Cid().String())
	}

	// Now make a bunch of dirs
	for i := 0; i < ndiriter; i++ {
		db, err := uio.NewDirectory(nd.DAG)
		if err != nil {
			t.Fatal(err)
		}
		for j := 0; j < 1+rand.Intn(10); j++ {
			name := fmt.Sprintf("child%d", j)

			err := db.AddChild(nd.Context(), name, nodes[rand.Intn(len(nodes))])
			if err != nil {
				t.Fatal(err)
			}
		}
		newdir, err := db.GetNode()
		if err != nil {
			t.Fatal(err)
		}

		err = nd.DAG.Add(nd.Context(), newdir)
		if err != nil {
			t.Fatal(err)
		}

		nodes = append(nodes, newdir)
		npaths := getPaths(t, nd, newdir.Cid().String(), newdir.(*dag.ProtoNode))
		paths = append(paths, npaths...)
	}

	// Now read a bunch, concurrently
	wg := sync.WaitGroup{}
	errs := make(chan error)

	for s := 0; s < 4; s++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for i := 0; i < 2000; i++ {
				item, err := path.NewPath(paths[rand.Intn(len(paths))])
				if err != nil {
					errs <- err
					continue
				}

				relpath := strings.Replace(item.String(), item.Namespace(), "", 1)
				fname := gopath.Join(mnt.Dir, relpath)

				rbuf, err := os.ReadFile(fname)
				if err != nil {
					errs <- err
				}

				// nd.Context() is never closed which leads to
				// hitting 8128 goroutine limit in go test -race mode
				ctx, cancelFunc := context.WithCancel(context.Background())

				read, err := api.Unixfs().Get(ctx, item)
				if err != nil {
					errs <- err
				}

				data, err := io.ReadAll(read.(files.File))
				if err != nil {
					errs <- err
				}

				cancelFunc()

				if !bytes.Equal(rbuf, data) {
					errs <- errors.New("incorrect read")
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errs)
	}()

	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
}

// Test writing a file and reading it back.
func TestIpfsBasicDirRead(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	nd, mnt := setupIpfsTest(t, nil)
	defer mnt.Close()

	// Make a 'file'
	fi, data := randObj(t, nd, 10000)

	// Make a directory and put that file in it
	db, err := uio.NewDirectory(nd.DAG)
	if err != nil {
		t.Fatal(err)
	}
	err = db.AddChild(nd.Context(), "actual", fi)
	if err != nil {
		t.Fatal(err)
	}

	d1nd, err := db.GetNode()
	if err != nil {
		t.Fatal(err)
	}

	err = nd.DAG.Add(nd.Context(), d1nd)
	if err != nil {
		t.Fatal(err)
	}

	dirname := gopath.Join(mnt.Dir, d1nd.Cid().String())
	fname := gopath.Join(dirname, "actual")
	rbuf, err := os.ReadFile(fname)
	if err != nil {
		t.Fatal(err)
	}

	dirents, err := os.ReadDir(dirname)
	if err != nil {
		t.Fatal(err)
	}
	if len(dirents) != 1 {
		t.Fatal("Bad directory entry count")
	}
	if dirents[0].Name() != "actual" {
		t.Fatal("Bad directory entry")
	}

	if !bytes.Equal(rbuf, data) {
		t.Fatal("Incorrect Read!")
	}
}

// Test to make sure the filesystem reports file sizes correctly.
func TestFileSizeReporting(t *testing.T) {
	if testing.Short() {
		t.SkipNow()
	}
	nd, mnt := setupIpfsTest(t, nil)
	defer mnt.Close()

	fi, data := randObj(t, nd, 10000)
	k := fi.Cid()

	fname := gopath.Join(mnt.Dir, k.String())

	finfo, err := os.Stat(fname)
	if err != nil {
		t.Fatal(err)
	}

	if finfo.Size() != int64(len(data)) {
		t.Fatal("Read incorrect size from stat!")
	}
}
