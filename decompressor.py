import sys
import os
import lzma

# Get the track file path from parameters
track_file_path = sys.argv[1]
track_dir_path = os.path.dirname(sys.argv[1])

# Read the track file
with open(track_file_path, 'rb') as track_file:
    track_data = bytearray(track_file.read())

# Find the start of the compressed section of the file
compression_start_bytes = b'\x5d\x00\x00\x02'
track_data = track_data[track_data.find(compression_start_bytes):]

# Compression header should contain uncompressed size as 64 bit integer, file has it as 32 bit integer
track_data[9:9] = b'\x00\x00\x00\x00'

# Decompress track data
track_data_uncompressed = lzma.LZMADecompressor().decompress(track_data)

# Save decompressed track
with open(os.path.join(track_dir_path, 'track_uncompressed.trk'), 'wb') as track_file_uncompressed:
    track_file_uncompressed.write(track_data_uncompressed)
