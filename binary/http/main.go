package main

import (
	userbinary "github.com/msyamsula/portofolio/domain/user/binary"
)

func main() {
	// appName := "backend"

	// // load env
	// godotenv.Load(".env")

	// // initialize instrumentation
	// telemetry.InitializeTelemetryTracing(appName, os.Getenv("JAEGER_HOST"))

	// // register prometheus metrics
	// prometheus.MustRegister(urlhttp.HashCounter)
	// prometheus.MustRegister(urlhttp.RedirectCounter)

	// pg, re := initDataLayer()

	// // create userHandler
	// urlHandler := initUrlHandler(pg, re)
	// graphHandler := initGraphHandler()
	// messageHandler := initMessageHandler(pg)
	// chatgptHandler := initChatGptHandler()

	// // create server routes
	// r := mux.NewRouter()
	// // message
	// r.HandleFunc("/message", messageHandler.ManageMesage)
	// // graph
	// r.HandleFunc("/graph/{algo}", http.HandlerFunc(graphHandler.InitGraph(http.HandlerFunc(graphHandler.Algorithm))))
	// // url
	// r.HandleFunc("/short", urlHandler.HashUrl)
	// r.HandleFunc("/{shortUrl}", urlHandler.RedirectShortUrl)
	// // chat gpt
	// r.HandleFunc("/code/review", chatgptHandler.CodeReview)

	// // cors option
	// c := cors.New(cors.Options{
	// 	AllowedOrigins:   []string{"*"},                              // Allow all origins (adjust for security)
	// 	AllowedMethods:   []string{"GET", "POST", "OPTIONS", "HEAD"}, // Allow methods
	// 	AllowedHeaders:   []string{"Content-Type"},                   // Allow headers
	// 	AllowCredentials: true,                                       // Allows credentials (cookies, authorization headers)
	// })
	// corsHandler := c.Handler(r)

	// // server handler
	// http.Handle("/", otelhttp.NewHandler(corsHandler, "")) // use otelhttp for telemetry
	// http.Handle("/metrics", promhttp.Handler())            // endpoint exporter, for prometheus scrapping

	// // server start
	// port, err := strconv.Atoi(os.Getenv("PORT"))
	// if err != nil {
	// 	log.Fatal("error in port format", err)
	// }
	// err = http.ListenAndServe(fmt.Sprintf("0.0.0.0:%d", port), nil)
	// if errors.Is(err, http.ErrServerClosed) {
	// 	fmt.Printf("server closed\n")
	// } else if err != nil {
	// 	fmt.Printf("error starting server: %s\n", err)
	// 	os.Exit(1)
	// }

	// run user binary
	userbinary.Run()
}
