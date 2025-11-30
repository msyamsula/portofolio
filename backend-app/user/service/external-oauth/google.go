package externaloauth

import (
	"context"
	"encoding/json"
	"log"

	"github.com/msyamsula/portofolio/backend-app/pkg/randomizer"
	"golang.org/x/oauth2"
)

type authService struct {
	randomizer        randomizer.Randomizer // randomizer use
	oauthConfigGoogle *oauth2.Config
}

func (g *authService) GetRedirectUrlGoogle(ctx context.Context, browserCookies string) (string, error) {

	var err error
	var state string
	state, err = g.randomizer.String()
	if err != nil {
		return "", err
	}

	return g.oauthConfigGoogle.AuthCodeURL(state), nil
}

func (g *authService) GetUserDataGoogle(ctx context.Context, browserCookies, state, code string) (UserData, error) {

	// allowed login
	var token *oauth2.Token
	var err error
	token, err = g.oauthConfigGoogle.Exchange(ctx, code)
	if err != nil {
		return UserData{}, err
	}

	client := g.oauthConfigGoogle.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		log.Println("Failed getting user info:", err)
		return UserData{}, err
	}
	defer resp.Body.Close()

	/*
		response format
				{
			  "id": "12345678901234567890",
			  "email": "user@example.com",
			  "verified_email": true,
			  "name": "John Doe",
			  "given_name": "John",
			  "family_name": "Doe",
			  "picture": "https://lh3.googleusercontent.com/a-/AOh14Gg...",
			  "locale": "en"
			}
	*/

	userData := UserData{}
	json.NewDecoder(resp.Body).Decode(&userData)

	return userData, nil
}

// var (
// 	oauthConfigGoogle = &oauth2.Config{
// 		ClientID:     "", // overwrite on New function
// 		ClientSecret: "", // overwrite on New function
// 		RedirectURL:  "", // overwrite on New function
// 		Endpoint:     google.Endpoint,
// 		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email"},
// 	}
// )

// func main() {
// 	r := gin.Default()

// 	r.GET("/", func(c *gin.Context) {
// 		c.String(http.StatusOK, "Welcome! Go to /login to authenticate with Google")
// 	})

// 	r.GET("/login", func(c *gin.Context) {
// 		url := oauthConfigGoogle.AuthCodeURL("test", oauth2.S256ChallengeOption("verifier"))
// 		c.Redirect(http.StatusTemporaryRedirect, url)
// 	})

// 	r.GET("/auth/google/callback", func(c *gin.Context) {
// 		state := c.Query("state")
// 		if state != "test" {
// 			log.Println("Invalid OAuth state")
// 			c.AbortWithStatus(http.StatusUnauthorized)
// 			return
// 		}

// 		code := c.Query("code")
// 		token, err := oauthConfigGoogle.Exchange(context.Background(), code, oauth2.VerifierOption("testing"))
// 		if err != nil {
// 			log.Println("Code exchange failed:", err)
// 			c.AbortWithStatus(http.StatusInternalServerError)
// 			return
// 		}

// 		client := oauthConfigGoogle.Client(context.Background(), token)
// 		resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
// 		if err != nil {
// 			log.Println("Failed getting user info:", err)
// 			c.AbortWithStatus(http.StatusInternalServerError)
// 			return
// 		}
// 		defer resp.Body.Close()

// 		userInfo := make([]byte, resp.ContentLength)
// 		resp.Body.Read(userInfo)

// 		c.String(http.StatusOK, fmt.Sprintf("User Info: %s", userInfo))
// 	})

// 	r.Run(":8080")
// }
