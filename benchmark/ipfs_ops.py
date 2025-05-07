import subprocess
import time

def add_file(file_path, chunker="size-262144"):
    start = time.time()
    output = subprocess.check_output(["ipfs", "add", "--chunker", chunker, "-Q", file_path])
    end = time.time()
    cid = output.decode().strip()
    return cid, end - start

def retrieve_file(cid):
    start = time.time()
    output = subprocess.check_output(["ipfs", "cat", cid])
    end = time.time()
    return len(output), end - start