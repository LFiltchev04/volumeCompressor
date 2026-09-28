package CephStorage

import(
	"io"
)

func ct(){
	//writer := io.WriteCloser(nil)
}


type LinkedWriter struct{
	pvPath string
	uname string

	wc io.WriteCloser
}

func (lw *LinkedWriter) Write(p []byte) (n int, err error) {
	
}


func (lw *LinkedWriter) Close() error {

}