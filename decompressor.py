import sys
import os
import lzma

# Get the track file path from parameters
track_file_path = sys.argv[1]
track_dir_path = os.path.dirname(sys.argv[1])

# Read the track file
with open(track_file_path, 'rb') as track_file:
    track_content = bytearray(track_file.read())

# Find the start of the compressed section of the file
compressed_start_bytes = b'\x5d\x00\x00\x02'
compressed_start_index = track_content.find(compressed_start_bytes)

# If index was found file is compressed
if compressed_start_index >= 0:
    # Separate track header and data
    track_header = track_content[:compressed_start_index]
    track_data = track_content[compressed_start_index:]

    # Compression header should contain uncompressed size as 64 bit integer, file has it as 32 bit integer
    track_data[9:9] = b'\x00\x00\x00\x00'

    # Decompress track data
    track_data_uncompressed = lzma.decompress(track_data)

    # Add track header back to the start of the file
    track_content_uncompressed = track_header + track_data_uncompressed

    # Save decompressed track data
    with open(os.path.join(track_dir_path, 'track_decompressed.trk'), 'wb') as track_file_uncompressed:
        track_file_uncompressed.write(track_content_uncompressed)

else:
    # Track file ends with the compression algorithm name
    header_end_bytes = b'LZMA'
    decompressed_start_index = track_content.find(header_end_bytes)

    if decompressed_start_index >= 0:
        # Get the index of the uncompressed track data start
        decompressed_start_index += len(header_end_bytes)

        # Separate track header and data
        track_header = track_content[:decompressed_start_index]
        track_data = track_content[decompressed_start_index:]

        # Compress track data
        compression_filters = [{'id': lzma.FILTER_LZMA1, 'dict_size': 0x0020000}]
        track_data_compressed = bytearray(lzma.compress(track_data, format=lzma.FORMAT_ALONE, filters=compression_filters))

        # Add uncompressed size to compression header
        track_uncompressed_len = len(track_data).to_bytes(length=4, byteorder='little')
        track_data_compressed[5:13] = track_uncompressed_len

        # Add track header back to the start of the file
        track_content_compressed = track_header + track_data_compressed

        # Save compressed track
        with open(os.path.join(track_dir_path, 'track_compressed.trk'), 'wb') as track_file_compressed:
            track_file_compressed.write(track_content_compressed)
