from fastcdc import fastcdc

def fastcdc_chunk(file_path, avg_size=262144, min_size=16384, max_size=1048576):
    with open(file_path, 'rb') as f:
        data = f.read()
    return [data[chunk.offset:chunk.offset+chunk.length] for chunk in fastcdc(data, min_size, avg_size, max_size)]