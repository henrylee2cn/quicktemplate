package quicktemplate

import (
	"testing"
)

func TestHTMLEscapeWriter(t *testing.T) {
	testHTMLEscapeWriter(t, "", "")
	testHTMLEscapeWriter(t, "foobar", "foobar")
	testHTMLEscapeWriter(t, `<h1>fo'"bar</h1>`, "&lt;h1&gt;fo&#39;&quot;bar&lt;/h1&gt;")
	testHTMLEscapeWriter(t, "fo<b>bar", "fo&lt;b&gt;bar")
}

func testHTMLEscapeWriter(t *testing.T, s, expectedS string) {
	bb := AcquireByteBuffer()
	w := acquireHTMLEscapeWriter(bb)
	n, err := w.Write([]byte(s))
	if err != nil {
		t.Fatalf("unexpected error when writing %q: %s", s, err)
	}
	if n != len(s) {
		t.Fatalf("unexpected n returned: %d. Expecting %d. s=%q", n, len(s), s)
	}
	releaseHTMLEscapeWriter(w)

	if string(bb.B) != expectedS {
		t.Fatalf("unexpected result: %q. Expecting %q", bb.B, expectedS)
	}
	ReleaseByteBuffer(bb)
}
