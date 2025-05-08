# Team Description

Project team: Debunking Filecoin


# Contribution


- **Yuxing Li (yuxingl5)**  
Implemented the benchmark pipeline for IPFS chunking.
Explored and tested lotus commands in a virtual machine environment to better understand Filecoin's storage and retrieval workflows

- **Yusen Wang (yusenw2)**  
Explored the IPFS (Kubo) source code to understand its internal chunking logic.
Modified the core logic responsible for the chunking algorithm, extended IPFS to support an alternative CDC for later benchmarking.

- **Jing Liu (jingl17)**
Explored the Filecoins documentations to understand PoRep & PoSt logic. Developed a simulation of the process to better understand the concept.
Actually deployed lotus and lotus-miner node with full sync to calibration network and local network then pledged a sector of 2KiB to test the functionality.

- **Benhao Lu (benhaol2)**  


# Overview

## Simulation of Filecoin

## Modifying the IPFS Protocol in Go

We explored and modified the Go implementation of the IPFS protocol, particularly focusing on how IPFS chunks and stores files. Our goal is to better understand the design trade-offs, performance bottlenecks, and real-world behavior of IPFS's chunking. 

We modified the Kubo (go-ipfs) implementation by replacing its default chunking strategy with a content-defined chunking (CDC) algorithm, specifically FastCDC, to analyze:
- The impact of chunking method on deduplication
- Differences in chunk boundary behavior
- Integration feasibility of external chunkers like those from boxo/chunker

We modified the Kubo codebase to plug in a custom FastCDC chunker implementation based on github.com/restic/chunker

### Implementation Highlights

We created a new chunker class under chunk/fastcdc.go

```go
package chunk

import (
	"io"
	"github.com/restic/chunker"
)

// implements the Splitter interface using FastCDC.
type FastCDC struct {
	ch     *chunker.Chunker
	reader io.Reader
}

func NewFastCDC(r io.Reader, avgBlkSize uint64) Splitter {
	const (
		MinSize = 512
		MaxSize = 64 * 1024
	)
	const poly = chunker.Pol(0x3DA3358B4DC173)

	ch := chunker.NewWithBoundaries(r, poly, MinSize, MaxSize)

	return &FastCDC{
		ch:     ch,
		reader: r,
	}
}

func (f *FastCDC) NextBytes() ([]byte, error) {
	chunk, err := f.ch.Next(nil)
	if err != nil {
		return nil, err
	}
	return chunk.Data, nil
}

func (f *FastCDC) Reader() io.Reader {
	return f.reader
}
```
### How to Run

```bash
$ cd kubo
$ go mod tidy       
$ go install ./cmd/ipfs

# Initialise and start a node
$ ~/go/bin/ipfs init
$ ~/go/bin/ipfs daemon &

# test with FastCDC
$ ~/go/bin/ipfs add --chunker=fastcdc-8192 myfile.dat
```
### 
