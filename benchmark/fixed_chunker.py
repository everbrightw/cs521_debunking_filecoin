
def fixed_chunk(file_path, chunk_size=262144):
    chunks = []
    with open(file_path, 'rb') as f:
        while chunk := f.read(chunk_size):
            chunks.append(chunk)
    return chunks
