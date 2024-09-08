package api

import "net/http"

func InitializeRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /create-room", CorsMiddleware(http.HandlerFunc(createRoomHandler)))
	mux.HandleFunc("/join-room/{id}", CorsMiddleware(http.HandlerFunc(joinRoomHandler)))
	mux.HandleFunc("GET /room-invitation/{id}", CorsMiddleware(http.HandlerFunc(setInvKeyHandler)))
	mux.HandleFunc("GET /sse-key-update/{id}", CorsMiddleware(http.HandlerFunc(sseKeyUpdateHandler)))
	mux.HandleFunc("/authorize-invite", CorsMiddleware(http.HandlerFunc(authorizeInvitationHandler)))
	mux.HandleFunc("/signalling/{id}", CorsMiddleware(http.HandlerFunc(signallingHandler)))

	mux.HandleFunc("/ws", CorsMiddleware(http.HandlerFunc(webSocketConnHandler)))

	return mux
}
