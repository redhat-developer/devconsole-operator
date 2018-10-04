// Code generated by vfsgen; DO NOT EDIT.

// +build !dev

package template

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	pathpkg "path"
	"time"
)

// Assets statically implements the virtual filesystem provided to vfsgen.
var Assets = func() http.FileSystem {
	fs := vfsgen۰FS{
		"/": &vfsgen۰DirInfo{
			name:    "/",
			modTime: time.Date(2018, 10, 4, 7, 6, 8, 650799297, time.UTC),
		},
		"/innerloop": &vfsgen۰DirInfo{
			name:    "innerloop",
			modTime: time.Date(2018, 10, 4, 14, 32, 14, 410087417, time.UTC),
		},
		"/innerloop/deploymentconfig": &vfsgen۰CompressedFileInfo{
			name:             "deploymentconfig",
			modTime:          time.Date(2018, 10, 4, 10, 33, 38, 501015106, time.UTC),
			uncompressedSize: 2080,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xbc\x55\x4d\x6f\xe3\x46\x0c\xbd\xfb\x57\x10\xb9\xcb\xda\x0d\xb0\x97\xc9\xc9\xb5\xdb\x22\x01\x92\x0a\x71\x9a\x6b\x40\x8f\x68\x99\xe9\x7c\x61\x66\xa4\xc2\x08\xf2\xdf\x8b\xb1\x1c\x79\x6c\xcb\x09\x5a\xa0\x3b\x27\x93\xf4\x23\x1f\xf9\x08\x0a\x1d\x3f\x93\x0f\x6c\x8d\x00\x74\x2e\x4c\xad\x23\x13\x36\xbc\x8e\x53\xb6\x65\xf7\x7d\xf2\x17\x9b\x5a\xc0\x82\x9c\xb2\x5b\x4d\x26\xce\xad\x59\x73\x33\xd1\x14\xb1\xc6\x88\x62\x02\xa0\x70\x45\x2a\xa4\x5f\x90\x72\x08\x78\x7b\x9b\x3e\xa0\xa6\xf7\xf7\x09\x80\x41\x4d\xb9\x27\x38\x92\xe9\xaf\x9e\x9c\x62\x89\x41\xc0\xf7\x09\x40\x20\x45\x32\x5a\x7f\x21\x09\x40\x3d\x10\x90\x3b\x02\xc7\xf1\x10\x3d\x46\x6a\xb6\x3d\x3c\x6e\x1d\x09\x78\xb4\x4a\xb1\x69\x26\x00\x91\xb4\x53\x18\xa9\x8f\xe6\xcc\xd3\xcb\xd9\x5f\x28\xfe\x35\x81\xf4\x4e\x3b\x4d\xbe\x8f\x6e\xd3\x93\xd6\x44\x64\x43\x7e\x28\x56\x00\xfa\x26\x2b\x5d\x40\x21\x33\xa3\xec\xd0\x97\x8a\x57\x65\x68\x1d\xf9\x8e\x83\xf5\x75\x99\xaa\x67\x8e\x69\xb2\x07\x8c\xb4\x5a\xa3\xa9\xc5\x57\x49\x56\x6c\x72\x7b\xf8\x3b\x99\x2e\xc7\xf6\x1d\xdd\xcd\x9e\x67\x2f\xb3\xaa\x7a\x59\xdc\x3e\x0e\x41\x80\x0e\x55\x4b\x02\xca\xc3\x60\xc2\x38\x74\xf1\xeb\x2f\x7f\xfe\x7e\x0e\xbc\x8a\xbe\xa5\xab\x4f\x20\x2f\xd5\x1f\x8f\x4f\x23\xb8\x1f\xdf\xbe\xfd\xb8\x80\x4b\x2c\xef\x66\x23\x2c\xd1\xb9\xe9\x2b\xfa\x21\xc0\x1a\x9b\x94\xab\xa6\xae\x08\xd7\x2c\xd2\x7a\x84\x78\x75\x1c\xaf\x5a\xa5\x2a\xab\x58\x6e\x05\xcc\xd4\xdf\xb8\x3d\x34\x38\xa6\x75\x7a\xce\xfa\x78\x24\xe8\x20\x7a\x65\x7d\xdc\x21\x96\x8e\xe4\x34\x59\x19\x0c\xc0\x79\x1b\xad\xb4\x4a\xc0\xd3\xbc\x1a\xfc\x9d\x55\xad\xa6\x7b\xdb\x9a\xe3\xac\x3a\x79\x2a\x8c\x1b\x31\xaa\x6f\x96\xb7\x27\x1a\x36\xe8\xa9\x2e\xd2\xda\x5f\xc8\x12\xb5\x2b\xd1\x47\x5e\xa3\xcc\x74\xfc\xc0\xeb\xeb\x1c\xcb\x86\xd3\x19\x38\x5b\xe6\xd1\xed\x99\xdf\x2f\x96\xe7\x7a\xf8\xd6\x14\xaf\xd8\xa1\x28\xdb\xe0\x4b\x65\x25\xaa\x32\x5c\x73\xe9\x5b\x73\x23\xad\x76\xac\x68\x34\x8e\x21\x90\x5e\x29\xba\x59\xb5\xac\x6a\x91\x2f\x5f\xb9\x73\xa1\x73\x67\x22\x7f\x0c\x7d\x79\x98\x50\x2f\xdb\x7f\x94\x5d\x5a\xb7\x2d\xc6\xc6\xed\x29\xd8\xd6\x4b\x0a\x02\xde\x0e\xe2\x46\xf2\x9a\x0d\x46\xb6\xe6\x9e\x42\x48\x15\xfa\x99\xd7\xd4\x95\x59\xb0\x50\xb6\xf9\x0c\xb4\xa7\xf4\x1b\x2b\xfa\x09\x0b\xd2\x67\xce\xd5\xd5\x2e\x6e\x17\xec\x8f\x7a\xbb\x84\x2e\x46\x17\x07\xc0\xa5\x6f\x4d\x88\x64\xe2\xf3\x2e\xff\x5c\x21\x6b\x91\xd1\x91\xc9\xf1\x70\x82\x8d\x9e\x9b\x66\xbf\x6a\xc5\xfe\xbc\xdf\x26\xa9\xe6\x1b\x34\x4d\x3f\x0d\x3e\xd8\x15\x7a\xd4\x03\x73\x6c\xa3\xd5\x18\x59\x0a\x48\x17\xe7\xf4\x18\xa7\x5a\x59\x97\x17\xb4\x5d\x7b\x9b\xd1\xec\xbf\x89\x3b\x06\xcb\xe8\x09\xf5\x13\x36\x5f\xec\xc8\x7e\xd5\xfe\xff\x06\x4e\x6f\xd2\xbf\x66\x7e\x7c\x12\xff\x09\x00\x00\xff\xff\xac\x2a\x5b\x5a\x20\x08\x00\x00"),
		},
		"/innerloop/imagestream": &vfsgen۰CompressedFileInfo{
			name:             "imagestream",
			modTime:          time.Date(2018, 10, 4, 14, 32, 14, 408825608, time.UTC),
			uncompressedSize: 679,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\x90\x4f\x8b\xdb\x30\x10\xc5\xef\xfa\x14\x8f\xa5\xd7\xd8\x6c\x8f\x5a\x7a\x58\xda\x4b\xa1\x94\xd2\x40\xef\x13\x79\x9c\xaa\xd1\x9f\x41\x92\x03\x41\xf8\xbb\x17\x3b\x8a\x37\xcb\xe6\x64\xf1\x66\xe6\xbd\x9f\x5f\xad\x9f\x02\x79\x7e\x15\x81\xfe\x82\xee\x27\x79\xc6\x3c\x93\xd8\x3f\x9c\xb2\x8d\x41\xe3\xfc\xac\x4e\x36\x0c\x1a\x3f\x6c\x2e\xca\x73\xa1\x81\x0a\x69\x05\x38\x3a\xb0\xcb\xcb\x0b\x20\x11\x8d\x37\xb3\x79\x56\xc0\xf2\x5e\xc4\xd5\x75\x9e\x95\x2d\xec\xb3\x56\xb5\x22\x51\x38\x32\xba\xbd\xb0\xe9\xbe\x7b\x3a\x2e\x99\x6a\x87\xfb\x58\xbb\xc8\x5d\x14\x0e\xf9\xaf\x1d\x4b\x67\x63\x7f\x7e\x56\xc0\x95\x65\x3d\xda\x97\xc4\xe4\x15\x70\x0f\xf5\x1e\xeb\x31\xd8\x47\x34\x20\x0b\x9b\x76\x1f\xe3\x69\x92\x5f\xd1\x59\x73\xb9\xb9\xb8\x68\xc8\x69\x8c\xe4\x32\xaf\x52\xad\x76\x44\xf7\x2d\x9a\x13\xa7\x15\xa6\x19\x17\x3a\xb6\xe8\x5d\xdb\x79\x0d\x21\x16\x2a\x36\x86\xaf\x7e\xc8\x6d\x0d\xa0\x4d\xde\x50\x01\xe3\x87\xac\xf1\x94\xa6\xb0\xfb\x47\x67\xd2\xfd\x94\x53\xbf\x66\xf7\xf9\xb3\xed\xd3\x14\x5e\x4c\xf4\x62\x1d\x3f\x9c\x53\xce\xec\x0f\x8e\x5f\x0e\x93\x75\x83\xee\x07\x16\x17\x2f\x9e\x43\xc9\xfd\x2a\x91\xc8\x53\x0b\xab\x95\xc3\xb0\xd1\x8c\x29\xfa\x37\x8c\x6b\xc7\x77\x3f\xb7\x4d\xb6\xde\x7e\xb3\xc4\xed\xda\x7a\x89\xa9\xb4\xc6\x50\x6f\xf2\x75\xd9\x51\xe1\x5c\x9a\x94\x78\xe4\xc4\xc1\xf0\xfb\x7a\x81\x72\x11\xd6\xd8\xc7\x29\x99\x5b\xc1\x2b\x5f\xfb\xfc\x0f\x00\x00\xff\xff\x5c\x1e\x56\x09\xa7\x02\x00\x00"),
		},
		"/innerloop/pvc": &vfsgen۰CompressedFileInfo{
			name:             "pvc",
			modTime:          time.Date(2018, 10, 4, 9, 16, 23, 850123135, time.UTC),
			uncompressedSize: 250,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x64\x8e\xbd\x8a\xc3\x30\x10\x84\x7b\x3d\xc5\xbe\xc0\x19\xae\x55\xeb\xfa\x8e\x03\x83\xfb\x3d\x79\x08\x22\xfa\x8b\x76\x1d\x08\x46\xef\x1e\xe4\x98\x10\x70\xb9\x33\xdf\xec\x0c\x17\x3f\xa3\x8a\xcf\xc9\xd2\xfd\xdb\x5c\x7d\x5a\x2c\xfd\x75\x45\x14\x49\xe7\x1c\xd6\x88\x31\xb0\x8f\x26\x42\x79\x61\x65\x6b\x88\x12\x47\x58\xda\xb6\x61\x2a\x70\xc3\xa4\xb9\xf2\x05\xc3\x2f\x47\xb4\x66\x88\x02\xff\x23\x48\x07\x89\xb8\x94\x9d\x7c\x9b\x1f\xe9\x43\x93\x02\xd7\x61\x76\x0e\x22\x3f\x79\xc1\x9e\xfd\x3a\x15\x74\x6b\xff\x51\x21\x79\xad\x0e\x47\x47\xc5\x6d\x85\xe8\x71\x11\xc9\x8b\x3f\x2f\x1c\xb9\xb0\xf3\xfa\x68\xcd\x3c\x03\x00\x00\xff\xff\x76\xbd\xed\x76\xfa\x00\x00\x00"),
		},
		"/innerloop/route": &vfsgen۰CompressedFileInfo{
			name:             "route",
			modTime:          time.Date(2018, 8, 14, 17, 14, 16, 0, time.UTC),
			uncompressedSize: 172,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x6c\xcd\x31\x0e\xc2\x30\x0c\x85\xe1\x3d\xa7\xf0\x09\x82\x58\x73\x08\x06\x90\xd8\x4d\xfb\x10\x16\x4d\x6c\x25\xa6\x4b\xd5\xbb\xa3\x28\x13\xa2\xf3\xff\x3e\x3d\x36\xb9\xa3\x36\xd1\x92\xa8\xea\xc7\x11\xd5\x50\xda\x4b\x9e\x1e\x45\x4f\xeb\x39\xbc\xa5\xcc\x89\xae\xbd\x85\x0c\xe7\x99\x9d\x53\x20\x2a\x9c\x91\x68\xdb\xe2\x85\x33\xf6\x3d\x10\x2d\xfc\xc0\xd2\x7a\x23\x62\xb3\xdf\xf8\x0f\x9a\x61\xea\x63\xd7\x41\xc6\xd1\x0d\x75\x95\x09\x47\xe2\x1b\x00\x00\xff\xff\xdd\xc7\x7e\xf9\xac\x00\x00\x00"),
		},
		"/innerloop/service": &vfsgen۰CompressedFileInfo{
			name:             "service",
			modTime:          time.Date(2018, 10, 4, 7, 33, 27, 665611911, time.UTC),
			uncompressedSize: 257,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x74\x8e\x31\x8a\xc3\x30\x10\x45\x7b\x9d\x62\x2e\xb0\x86\x6d\xd5\x6e\xbf\x18\xbc\x6c\x3f\x91\x7f\x8c\x88\xa4\x19\xa4\xc1\x10\x8c\xef\x1e\x2c\x52\x24\x24\xe9\x86\x79\xef\xc1\x67\x8d\xff\xa8\x2d\x4a\xf1\xb4\x7e\xbb\x4b\x2c\xb3\xa7\x09\x75\x8d\x01\x2e\xc3\x78\x66\x63\xef\x88\x0a\x67\x78\xda\xb6\xe1\x97\x33\xf6\xdd\x11\x25\x3e\x21\xb5\x83\x11\xb1\xea\x33\x7c\x0d\x9a\x22\x1c\xb2\x4a\xb5\x5e\x7d\xf5\xb3\x2b\x93\x22\x0c\xa3\x54\xbb\xb7\x5a\xc5\x24\x48\xf2\xf4\xf7\x33\xf6\x8f\x71\x5d\x60\xe3\xfb\xa0\x21\x21\x98\xd4\x8f\x5b\x66\x68\x92\x6b\x46\xb1\x20\xe5\x1c\x97\x07\x7e\x0b\x00\x00\xff\xff\x7c\x58\xa5\x67\x01\x01\x00\x00"),
		},
	}
	fs["/"].(*vfsgen۰DirInfo).entries = []os.FileInfo{
		fs["/innerloop"].(os.FileInfo),
	}
	fs["/innerloop"].(*vfsgen۰DirInfo).entries = []os.FileInfo{
		fs["/innerloop/deploymentconfig"].(os.FileInfo),
		fs["/innerloop/imagestream"].(os.FileInfo),
		fs["/innerloop/pvc"].(os.FileInfo),
		fs["/innerloop/route"].(os.FileInfo),
		fs["/innerloop/service"].(os.FileInfo),
	}

	return fs
}()

type vfsgen۰FS map[string]interface{}

func (fs vfsgen۰FS) Open(path string) (http.File, error) {
	path = pathpkg.Clean("/" + path)
	f, ok := fs[path]
	if !ok {
		return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
	}

	switch f := f.(type) {
	case *vfsgen۰CompressedFileInfo:
		gr, err := gzip.NewReader(bytes.NewReader(f.compressedContent))
		if err != nil {
			// This should never happen because we generate the gzip bytes such that they are always valid.
			panic("unexpected error reading own gzip compressed bytes: " + err.Error())
		}
		return &vfsgen۰CompressedFile{
			vfsgen۰CompressedFileInfo: f,
			gr: gr,
		}, nil
	case *vfsgen۰DirInfo:
		return &vfsgen۰Dir{
			vfsgen۰DirInfo: f,
		}, nil
	default:
		// This should never happen because we generate only the above types.
		panic(fmt.Sprintf("unexpected type %T", f))
	}
}

// vfsgen۰CompressedFileInfo is a static definition of a gzip compressed file.
type vfsgen۰CompressedFileInfo struct {
	name              string
	modTime           time.Time
	compressedContent []byte
	uncompressedSize  int64
}

func (f *vfsgen۰CompressedFileInfo) Readdir(count int) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("cannot Readdir from file %s", f.name)
}
func (f *vfsgen۰CompressedFileInfo) Stat() (os.FileInfo, error) { return f, nil }

func (f *vfsgen۰CompressedFileInfo) GzipBytes() []byte {
	return f.compressedContent
}

func (f *vfsgen۰CompressedFileInfo) Name() string       { return f.name }
func (f *vfsgen۰CompressedFileInfo) Size() int64        { return f.uncompressedSize }
func (f *vfsgen۰CompressedFileInfo) Mode() os.FileMode  { return 0444 }
func (f *vfsgen۰CompressedFileInfo) ModTime() time.Time { return f.modTime }
func (f *vfsgen۰CompressedFileInfo) IsDir() bool        { return false }
func (f *vfsgen۰CompressedFileInfo) Sys() interface{}   { return nil }

// vfsgen۰CompressedFile is an opened compressedFile instance.
type vfsgen۰CompressedFile struct {
	*vfsgen۰CompressedFileInfo
	gr      *gzip.Reader
	grPos   int64 // Actual gr uncompressed position.
	seekPos int64 // Seek uncompressed position.
}

func (f *vfsgen۰CompressedFile) Read(p []byte) (n int, err error) {
	if f.grPos > f.seekPos {
		// Rewind to beginning.
		err = f.gr.Reset(bytes.NewReader(f.compressedContent))
		if err != nil {
			return 0, err
		}
		f.grPos = 0
	}
	if f.grPos < f.seekPos {
		// Fast-forward.
		_, err = io.CopyN(ioutil.Discard, f.gr, f.seekPos-f.grPos)
		if err != nil {
			return 0, err
		}
		f.grPos = f.seekPos
	}
	n, err = f.gr.Read(p)
	f.grPos += int64(n)
	f.seekPos = f.grPos
	return n, err
}
func (f *vfsgen۰CompressedFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		f.seekPos = 0 + offset
	case io.SeekCurrent:
		f.seekPos += offset
	case io.SeekEnd:
		f.seekPos = f.uncompressedSize + offset
	default:
		panic(fmt.Errorf("invalid whence value: %v", whence))
	}
	return f.seekPos, nil
}
func (f *vfsgen۰CompressedFile) Close() error {
	return f.gr.Close()
}

// vfsgen۰DirInfo is a static definition of a directory.
type vfsgen۰DirInfo struct {
	name    string
	modTime time.Time
	entries []os.FileInfo
}

func (d *vfsgen۰DirInfo) Read([]byte) (int, error) {
	return 0, fmt.Errorf("cannot Read from directory %s", d.name)
}
func (d *vfsgen۰DirInfo) Close() error               { return nil }
func (d *vfsgen۰DirInfo) Stat() (os.FileInfo, error) { return d, nil }

func (d *vfsgen۰DirInfo) Name() string       { return d.name }
func (d *vfsgen۰DirInfo) Size() int64        { return 0 }
func (d *vfsgen۰DirInfo) Mode() os.FileMode  { return 0755 | os.ModeDir }
func (d *vfsgen۰DirInfo) ModTime() time.Time { return d.modTime }
func (d *vfsgen۰DirInfo) IsDir() bool        { return true }
func (d *vfsgen۰DirInfo) Sys() interface{}   { return nil }

// vfsgen۰Dir is an opened dir instance.
type vfsgen۰Dir struct {
	*vfsgen۰DirInfo
	pos int // Position within entries for Seek and Readdir.
}

func (d *vfsgen۰Dir) Seek(offset int64, whence int) (int64, error) {
	if offset == 0 && whence == io.SeekStart {
		d.pos = 0
		return 0, nil
	}
	return 0, fmt.Errorf("unsupported Seek in directory %s", d.name)
}

func (d *vfsgen۰Dir) Readdir(count int) ([]os.FileInfo, error) {
	if d.pos >= len(d.entries) && count > 0 {
		return nil, io.EOF
	}
	if count <= 0 || count > len(d.entries)-d.pos {
		count = len(d.entries) - d.pos
	}
	e := d.entries[d.pos : d.pos+count]
	d.pos += count
	return e, nil
}
