package stick

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFilesystemLoader(t *testing.T) {
	d, _ := os.Getwd()
	l := NewFilesystemLoader(d)
	f, e := l.Load("testdata/base.txt.twig")
	if e != nil {
		t.Errorf("expected load to succeed. %s", e)
	} else if f.Name() != "testdata/base.txt.twig" {
		t.Errorf("unexpected template name: %s", f.Name())
	}

	_, e = l.Load("testdata/doesnt_exists.txt.twig")
	if e == nil {
		t.Error("expected error, got nil")
	} else if !os.IsNotExist(e) {
		t.Errorf("expected os.NotExist error, got %s", e)
	}
}

func TestStringLoader(t *testing.T) {
	l := &StringLoader{}
	b, e := l.Load("test string")
	if e != nil {
		t.Errorf("expected load to succeed got %s", e)
	} else if b.Name() != "test string" {
		t.Errorf("unexpected template name: %s", b.Name())
	}
}

func TestMemoryLoader(t *testing.T) {
	l := &MemoryLoader{map[string]string{"test.twig": "some text"}}
	b, e := l.Load("test.twig")
	if e != nil {
		t.Fatalf("expected load to succeed got %s", e)
	} else if b.Name() != "test.twig" {
		t.Fatalf("expected to load test.twig got %s", b.Name())
	}
	s, e := ioutil.ReadAll(b.Contents())
	if e != nil {
		t.Fatalf("unexpected error %s", e)
	}
	if string(s) != "some text" {
		t.Fatalf("expected 'some text' got '%s'", string(s))
	}
}
