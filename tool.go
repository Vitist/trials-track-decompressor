package main

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ulikunitz/xz/lzma"
)

// Magic bytes for detecting different sections of the file
var (
	compressedStartString  = []byte("LZMA")                       // Fusion and Rising file compression starts with this string
	compressedStartBytes   = []byte{0x5d, 0x00, 0x00, 0x02, 0x00} // Evo file compression starts with these compression properties
	trackHeaderEnd         = []byte{0x48, 0x45, 0x4E, 0x44, 0x00} // Header end bytes for detecting where the original track header ends
	uncompressedSizeFiller = []byte{0x00, 0x00, 0x00, 0x00}       // Header is supposed to have uncompressed size as 64 bit integer but file has 32 bit integer
)

// Limit for how far to the file to search for the magic bytes
const maxSearchIndex = 200

type gameFile struct {
	name   string
	header []byte
	data   []byte
}

func (gf gameFile) decompress() error {
	buf := bytes.NewBuffer(gf.data)
	r, err := lzma.NewReader(buf)
	if err != nil {
		return err
	}

	dst, err := os.Create(gf.getDecompressedName())
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, r); err != nil {
		return err
	}

	return nil
}

func (gf gameFile) getDecompressedName() string {
	if len(gf.name) == 0 {
		return "decompressed"
	}

	nameParts := strings.Split(gf.name, ".")
	if len(nameParts) == 1 {
		return nameParts[0] + "_decompressed"
	} else {
		return nameParts[0] + "_decompressed." + nameParts[1]
	}
}

func readGameFile(path string) (gameFile, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return gameFile{}, err
	}

	name := filepath.Base(path)

	if index := bytes.Index(content[:maxSearchIndex], compressedStartString); index >= -1 {
		header := content[:index+len(compressedStartString)]
		compressed := content[index+len(compressedStartString):]
		compressed = append(compressed[:9], append(uncompressedSizeFiller, compressed[9:]...)...)
		return gameFile{name: name, header: header, data: compressed}, nil
	} else if index := bytes.Index(content[:maxSearchIndex], compressedStartBytes); index >= -1 {
		header := content[:index]
		compressed := content[index:]
		compressed = append(compressed[:9], append(uncompressedSizeFiller, compressed[9:]...)...)
		return gameFile{name: name, header: header, data: compressed}, nil
	}

	return gameFile{}, errors.New("Could not find start of compressed data")
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("No file path provided as argument")
	}

	compressedFile, err := readGameFile(os.Args[1])
	if err != nil {
		log.Fatalf("Error reading file: %s", err)
	}

	compressedFile.decompress()
	/*f, err := os.Open("a.lzma")
	if err != nil {
		log.Fatal(err)
	}
	// no need for defer; Fatal calls os.Exit(1) that doesn't execute deferred functions
	r, err := lzma.NewReader(bufio.NewReader(f))
	if err != nil {
		log.Fatal(err)
	}
	_, err = io.Copy(os.Stdout, r)
	if err != nil {
		log.Fatal(err)
	}
	if err := f.Close(); err != nil {
		log.Fatal(err)
	}*/
}

/* OLD PYTHON CODE
import sys
import os
import lzma

track_dir_path = os.path.dirname(sys.argv[1])

# Read the track file
with open(track_file_path, 'rb') as track_file:
    track_content = bytearray(track_file.read())

start_bytes_max_index = 200
if len(track_content) < start_bytes_max_index:
    start_bytes_max_index = len(track_content)

# Find the start of the compressed section of the file
# In Fusion and Rising files, compressed data section starts with the string LZMA followed by compression settings
compressed_start_string = b'LZMA'
# In Evolution files, compressed data sections starts with compression settings. The compression settings seem to be
# the same in all files.
compressed_start_bytes = b'\x5d\x00\x00\x02\x00'
compressed_start_index = track_content[:200].find(compressed_start_string)

if compressed_start_index != -1:
    # Fusion or Rising
    compressed_start_index += 4
else:
    # Evolution
    compressed_start_index = track_content[:200].find(compressed_start_bytes)

# Header end bytes for detecting where the original track header ends
track_header_end = b'\x48\x45\x4E\x44\x00'

# If index was found file is compressed
if compressed_start_index >= 0:
    print(compressed_start_index)
    print(track_content[compressed_start_index:compressed_start_index + 5])
    # Separate track header and data
    track_header = track_content[:compressed_start_index]
    track_header += track_header_end
    track_data = track_content[compressed_start_index:]

    # Compression header should contain uncompressed size as 64 bit integer, file has it as 32 bit integer
    track_data[9:9] = b'\x00\x00\x00\x00'

    # Decompress track data
    #b'\xdb\x00\x02\x00\x00'
    #b'\x5d\x00\x00\x02\x00'
    prop = 201
    if (prop > (4 * 5 + 4) * 9 + 8):
        print("Invalid?")

    pb = int(prop / (9 * 5))
    prop -= int(pb * 9 * 5)
    lp = int(prop / 9)
    lc = int(prop - lp * 9)
    print(f'{lc}/{lp}/{pb}')
    filter = lzma._decode_filter_properties(lzma.FILTER_LZMA1, b'\xc9\x00\x02\x00\x00')
    decompressor = lzma.LZMADecompressor(lzma.FORMAT_ALONE, None,[filter])
    track_data_uncompressed = decompressor.decompress(track_data)
    #track_data_uncompressed = lzma.decompress(track_data)

    # Add track header back to the start of the file
    track_content_uncompressed = track_header + track_data_uncompressed

    # Save decompressed track data
    with open(os.path.join(track_dir_path, 'track_decompressed.trk'), 'wb') as track_file_uncompressed:
        track_file_uncompressed.write(track_content_uncompressed)

else:
    # Track file header ends with the compression algorithm name
    decompressed_start_index = track_content[:200].find(track_header_end)

    if decompressed_start_index >= 0:
        # Get the index of the uncompressed track data start
        decompressed_start_index += len(track_header_end)

        # Separate track header and data
        track_header = track_content[:decompressed_start_index - len(track_header_end)]
        track_data = track_content[decompressed_start_index:]

        # Compress track data
        compression_filters = [{'id': lzma.FILTER_LZMA1, 'dict_size': 0x0020000}]
        track_data_compressed = bytearray(lzma.compress(track_data,
                                                        format=lzma.FORMAT_ALONE,
                                                        filters=compression_filters))

        # Add uncompressed size to compression header
        track_uncompressed_len = len(track_data).to_bytes(length=4, byteorder='little')
        track_data_compressed[5:13] = track_uncompressed_len

        # Add track header back to the start of the file
        track_content_compressed = track_header + track_data_compressed

        # Save compressed track
        with open(os.path.join(track_dir_path, 'track_compressed.trk'), 'wb') as track_file_compressed:
            track_file_compressed.write(track_content_compressed)

*/
