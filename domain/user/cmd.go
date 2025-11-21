package binary

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/msyamsula/portofolio/binary/postgres"
	userhttp "github.com/msyamsula/portofolio/domain/user/http"
	useroauth "github.com/msyamsula/portofolio/domain/user/oauth"
	"github.com/msyamsula/portofolio/domain/user/repository"
	"github.com/msyamsula/portofolio/domain/user/service"
)

func initUserHandler(userSvc *service.Service) *userhttp.Handler {

	handler := userhttp.New(userhttp.Dependencies{
		Service: userSvc,
	})
	return handler

}

func initDataLayer() *postgres.Postgres {

	pg := postgres.New(postgres.Config{
		Username: os.Getenv("POSTGRES_PASSWORD"),
		Password: os.Getenv("POSTGRES_USERNAME"),
		DbName:   os.Getenv("POSTGRES_DB"),
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
	})

	return pg
}

func initGoogleSigninService(userSvc *service.Service) *useroauth.Service {
	return useroauth.New(useroauth.Dependencies{
		GoogleClientId:      os.Getenv("GOOGLE_CLIENT_ID"),
		GoogleRedirectOauth: os.Getenv("GOOGLE_REDIRECT_OAUTH"),
		GoogleSecret:        os.Getenv("GOOGLE_SECRET"),
		UserSvc:             userSvc,
		RedirectChat:        os.Getenv("REDIRECT_CHAT"),
		OauthStateLength:    25,
		OauthCharacters:     os.Getenv("HASHER_CHARACTER_POOL"),
	})
}

func Run(r *mux.Router) {

	// load env
	godotenv.Load(".env")

	pg := initDataLayer()

	userSvc := service.New(service.Dependencies{
		Persistence: &repository.Persistence{
			Postgres: pg,
		},
	})

	// create userHandler
	userHandler := initUserHandler(userSvc)
	googleUserOauthHandler := initGoogleSigninService(userSvc)

	r.HandleFunc("/check", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("all good"))
	})

	// google sign in
	r.HandleFunc("/access/token", googleUserOauthHandler.HandleCallback)
	r.HandleFunc("/google/signin", googleUserOauthHandler.HandleLogin)
	// user
	r.HandleFunc("/user/friend", userHandler.ManageFriend)
	r.HandleFunc("/user", userHandler.ManageUser)
}
