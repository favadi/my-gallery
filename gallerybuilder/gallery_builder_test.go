package gallerybuilder

import "testing"

type mockStorage struct {
}

func (m *mockStorage) Store(_ []Image) error {
	return nil
}

func TestGallery_Build(t *testing.T) {
	g := NewGallery(&mockStorage{}, "testdata")
	images, err := g.lookupImages()
	if err != nil {
		t.Fatal(err)
	}
	// the text file should be ignored
	if len(images) != 2 {
		t.Errorf("wrong number of expected images: expected=%d actual=%d", 2, len(images))
	}
	t.Log(images)
}
