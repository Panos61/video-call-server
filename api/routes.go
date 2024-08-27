package api

import "net/http"

func InitializeRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /create-room", CorsMiddleware(http.HandlerFunc(createRoomHandler)))
	mux.HandleFunc("/join-room/{id}", CorsMiddleware(http.HandlerFunc(joinRoomHandler)))
	mux.HandleFunc("/grant-access", CorsMiddleware(http.HandlerFunc(grantRoomAccessHandler)))
	mux.HandleFunc("/signalling/{id}", CorsMiddleware(http.HandlerFunc(signallingHandler)))
	mux.HandleFunc("/ws/{id}/generate-key", updateInvKeyHandler)

	mux.HandleFunc("/ws", CorsMiddleware(http.HandlerFunc(webSocketConnHandler)))

	return mux
}
