import csv
import os
from ipfs_ops import add_file, retrieve_file

# files = ["sample_files/sample.pdf", "sample_files/sample.mp4"]
files = ["sample_files/sample.pdf"]
chunkers = [("size-262144", "Fixed"), ("rabin", "FastCDC")]

with open("results/benchmark_log.csv", "w", newline="") as csvfile:
    writer = csv.writer(csvfile)
    writer.writerow(["File", "Chunker", "UploadTime", "DownloadTime", "CID"])

    for f in files:
        for chunker, label in chunkers:
            cid, up_time = add_file(f, chunker=chunker)
            size, down_time = retrieve_file(cid)
            writer.writerow([f, label, up_time, down_time, cid])
            print(f"File: {f}, Chunker: {label}, Upload: {up_time:.2f}s, Download: {down_time:.2f}s, CID: {cid}")