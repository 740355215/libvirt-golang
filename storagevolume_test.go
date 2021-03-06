package libvirt

import (
	"bytes"
	"io"
	"testing"

	"github.com/cd1/utils-golang"
)

const deltaResizeChunkSize = 1024 // 1 KiB

func TestStorageVolumeInit(t *testing.T) {
	env := newTestEnvironment(t).withStorageVolume()
	defer env.cleanUp()

	key, err := env.vol.Key()
	if err != nil {
		t.Error(err)
	}

	if l := len(key); l == 0 {
		t.Error("empty storage volume key")
	}

	name, err := env.vol.Name()
	if err != nil {
		t.Error(err)
	}

	if name != env.volData.Name {
		t.Errorf("unexpected storage volume name; got=%v, want=%v", name, env.volData.Name)
	}

	path, err := env.vol.Path()
	if err != nil {
		t.Error(err)
	}

	if l := len(path); l == 0 {
		t.Error("empty storage volume path")
	}

	xml, err := env.vol.XML()
	if err != nil {
		t.Error(err)
	}

	if l := len(xml); l == 0 {
		t.Error("empty XML descriptor")
	}

	typ, err := env.vol.InfoType()
	if err != nil {
		t.Error(err)
	}

	if typ != VolTypeFile {
		t.Errorf("unexpected storage volume type; got=%v want=%v", typ, VolTypeFile)
	}

	_, err = env.vol.InfoCapacity()
	if err != nil {
		t.Error(err)
	}

	_, err = env.vol.InfoAllocation()
	if err != nil {
		t.Error(err)
	}
}

func TestStorageVolumeResize(t *testing.T) {
	env := newTestEnvironment(t).withStorageVolume()
	defer env.cleanUp()

	if err := env.vol.Resize(0, StorageVolumeResizeFlag(^uint32(0))); err == nil {
		t.Error("an error was not returned when resizing a storage volume with an invalid flag")
	}

	poolAvailable, err := env.pool.InfoAvailable()
	if err != nil {
		t.Fatal(err)
	}

	newSize := env.volData.Capacity
	if poolAvailable > deltaResizeChunkSize {
		newSize += deltaResizeChunkSize
	} else {
		newSize += poolAvailable
	}

	if err := env.vol.Resize(newSize, VolResizeDefault); err != nil {
		t.Error(err)
	}
}

func TestStorageVolumeWipe(t *testing.T) {
	env := newTestEnvironment(t).withStorageVolume()
	defer env.cleanUp()

	if err := env.vol.Wipe(StorageVolumeWipeAlgorithm(^uint32(0))); err == nil {
		t.Error("an error was not returned when wiping a storage volume with an invalid algorithm")
	}

	if err := env.vol.Wipe(VolWipeAlgZero); err != nil {
		t.Error(err)
	}
}

func TestStorageVolumeRef(t *testing.T) {
	env := newTestEnvironment(t).withStorageVolume()
	defer env.cleanUp()

	if err := env.vol.Ref(); err != nil {
		t.Fatal(err)
	}

	if err := env.vol.Free(); err != nil {
		t.Error(err)
	}
}

func TestStorageVolumeLookupPool(t *testing.T) {
	env := newTestEnvironment(t).withStorageVolume()
	defer env.cleanUp()

	pool, err := env.vol.StoragePool()
	if err != nil {
		t.Fatal(err)
	}

	uuid, err := pool.UUID()
	if err != nil {
		t.Fatal(err)
	}

	if uuid != env.poolData.UUID {
		t.Errorf("unexpected pool looked up from storage volume; got=%v, want=%v", uuid, env.poolData.UUID)
	}
}

func TestStorageVolumeUploadDownload(t *testing.T) {
	env := newTestEnvironment(t).withStorageVolume().withStream()
	defer env.cleanUp()

	if err := env.vol.Upload(Stream{}, 0, 0); err == nil {
		t.Error("an error was not returned when trying to set up an upload with an invalid stream")
	}

	if err := env.vol.Download(Stream{}, 0, 0); err == nil {
		t.Error("an error was not returned when trying to set up a download with an invalid stream")
	}

	data := utils.RandomString()
	dataLen := len(data)

	if err := env.vol.Upload(*env.str, 0, uint64(dataLen)); err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBufferString(data)

	nBytes, err := io.Copy(env.str, buf)
	if err != nil {
		t.Fatal(err)
	}

	if nBytes != int64(dataLen) {
		t.Errorf("unexpected number of bytes written; got=%v, want=%v", nBytes, dataLen)
	}

	if err = env.str.Finish(); err != nil {
		t.Fatal(err)
	}

	if err = env.vol.Download(*env.str, 0, uint64(dataLen)); err != nil {
		t.Fatal(err)
	}

	buf.Reset()

	nBytes, err = io.Copy(buf, env.str)
	if err != nil {
		t.Fatal(err)
	}

	if nBytes != int64(dataLen) {
		t.Errorf("unexpected number of bytes read; got=%v, want=%v", nBytes, dataLen)
	}

	if err = env.str.Finish(); err != nil {
		t.Fatal(err)
	}

	if dl := buf.String(); dl != data {
		t.Errorf("unexpected downloaded content; got=%v, want=%v", dl, data)
	}
}

func BenchmarkStorageVolumeResize(b *testing.B) {
	env := newTestEnvironment(b).withStorageVolume()
	defer env.cleanUp()

	poolAvailable, err := env.pool.InfoAvailable()
	if err != nil {
		b.Fatal(err)
	}

	if poolAvailable < uint64(b.N*deltaResizeChunkSize) {
		b.Skip("no available space on storage pool to perform benchmark")
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if err := env.vol.Resize(deltaResizeChunkSize, VolResizeDelta); err != nil {
			b.Error(err)
		}
	}
	b.StopTimer()
}

func BenchmarkStorageVolumeWipe(b *testing.B) {
	env := newTestEnvironment(b).withStorageVolume()
	defer env.cleanUp()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if err := env.vol.Wipe(VolWipeAlgZero); err != nil {
			b.Error(err)
		}
	}
	b.StopTimer()
}
