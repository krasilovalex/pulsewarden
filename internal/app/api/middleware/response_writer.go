package middleware

import "net/http"

type responseTracker struct {
	http.ResponseWriter

	wroteHeader  bool
	statusCode   int
	bytesWritten int
}

func trackResponse(w http.ResponseWriter) *responseTracker {
	if tracker, ok := w.(*responseTracker); ok {
		return tracker
	}

	return &responseTracker{
		ResponseWriter: w,
	}
}

func (w *responseTracker) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}

	w.wroteHeader = true
	w.statusCode = statusCode

	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseTracker) Write(data []byte) (int, error) {
	if !w.wroteHeader {
		w.WriteHeader(http.StatusOK)
	}

	written, err := w.ResponseWriter.Write(data)
	w.bytesWritten += written

	return written, err
}

func (w *responseTracker) StatusCode() int {
	if !w.wroteHeader {
		return http.StatusOK
	}

	return w.statusCode
}

func (w *responseTracker) BytesWritten() int {
	return w.bytesWritten
}

// UNWRAP ALLOWS http.ReponseController to reach the original ReponseWriter
func (w *responseTracker) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
