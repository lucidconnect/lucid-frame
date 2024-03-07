package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"lucid.frame/frame"
	"lucid.frame/neynar"
)

// nft item id can be in the url
// us the item id to c

// TODO:
// -  extract frame id from url
// -  fetch frame details and return a frame with given details
// -  parse frame action and execute action
// abigen --abi abi/lucidNft.abi --bin abi/lucidNft.bin --pkg lucidNft --type lucidNft --out ./lucidNft/lucidNft.go

var DB *gorm.DB

func main() {
	godotenv.Load()

	DB, _ = SetupDatabase()

	port := os.Getenv("PORT")
	r := mux.NewRouter()
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, fmt.Sprintln("Frame Server"))
	})
	loadCORS(r)
	r.HandleFunc("/frame/{frame}", frameHandler())
	r.HandleFunc("/createframe", createFrameHandler())
	r.HandleFunc("/getframe", fetchFrameHandler())
	fmt.Printf("Lucid frame server starting on port %v \n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func loadCORS(router *mux.Router) {
	allowedOrigins := []string{"*", "localhost:3000", "https://lucid-v2.vercel.app/"}
	c := cors.New(cors.Options{
		AllowedOrigins: allowedOrigins,
		AllowedMethods: []string{
			http.MethodOptions,
			http.MethodGet,
			http.MethodPost,
		},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: false,
	})
	router.Use(c.Handler)
}

func returnFrame(w http.ResponseWriter, frameId, imageUrl, txUrl string, buttons []frame.Button) {
	w.Header().Set("Content-Type", "text/html")
	// Write the HTML meta tags to the response
	// frameBtn := frame.Button(title)

	// var btns []frame.Button

	// btns = append(btns, frameBtn)

	ogFrame := frame.ParseFrame(imageUrl, frameId, txUrl, buttons...)
	fmt.Println(ogFrame)
	fmt.Fprint(w, ogFrame)
}

func frameHandler() http.HandlerFunc {
	type reqBody struct {
		UntrustedData map[string]any    `json:"untrustedData"`
		TrustedData   map[string]string `json:"trustedData"`
	}
	neynarApiKey := os.Getenv("NEYNAR_API_KEY")
	return func(w http.ResponseWriter, r *http.Request) {
		neynarClient, err := neynar.NewNeynarClient(neynar.WithApiKey(neynarApiKey))
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("an unexpected error occured"))
			return
		}

		vars := mux.Vars(r)
		frmaeId, ok := vars["frame"]
		if !ok {
			fmt.Println("id is missing in parameters")
		}

		frameDetails, err := frame.GetFrameDetails(frmaeId, DB)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("an unexpected error occured"))
			return
		}
		imageUrl := frameDetails.ImageUrl
		item := frameDetails.ItemId

		switch r.Method {
		case http.MethodGet:
			frameBtn := frame.ClaimButton

			var btns []frame.Button

			btns = append(btns, frameBtn)
			returnFrame(w, frmaeId, imageUrl, "", btns)
		case http.MethodPost:
			frameReqBody := reqBody{}
			err := json.NewDecoder(r.Body).Decode(&frameReqBody)
			if err != nil {
				log.Printf("error decoding request body %v", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("an unexpected error occured"))
				return
			}

			msgData := frameReqBody.TrustedData["messageBytes"]

			type validationBody struct {
				CastReactionContext bool   `json:"cast_reaction_context"`
				FollowContext       bool   `json:"follow_context"`
				MessageBytesInHex   string `json:"message_bytes_in_hex"`
			}

			vBody := validationBody{true, false, msgData}
			msgDataBytes, err := json.Marshal(vBody)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("an unexpected error occured"))
				return
			}

			action, err := neynarClient.ValidateFrameMessage(msgDataBytes)
			if err != nil {
				log.Printf("error validating frame %v", err)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("invalid frame message"))
				return
			}

			if len(action.Interactor.VerifiedAdresses.EthAddresses) == 0 {
				err := neynar.ErrNoVerifications
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			verifiedEthAddress := action.Interactor.VerifiedAdresses.EthAddresses[0]
			buttonTitle := action.Cast.Frames[0].Buttons[0].Title
			// buttonTitle := "claim"
			uv := r.URL.Query()
			claimed := uv.Get("claimed")
			txHash := uv.Get("tx")
			if claimed == "true" {
				buttonTitle = "make your own @"
			}
			if txHash != "" {
				buttonTitle = "view tx"
			}
			button := frame.Button(buttonTitle)
			buttonIdx := action.TappedButton.Index
			fmt.Printf("button %v clicked \n", buttonIdx)
			switch buttonIdx {
			case 1:
				// claim
				fmt.Println("Button: ", button)
				response, err := frame.ParseFrameAction(button, item, verifiedEthAddress)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				fmt.Println("Response: ", response)
				if button == frame.TransactionButton {
					redirect := fmt.Sprintf("%v/tx/%v", txHash)
					http.Redirect(w, r, redirect, http.StatusFound)
					return
				}
				image := "https://arweave.net/zTVSCzHxGyqWv9J5ZBwsHlyJ0ZNfM2SyANAnfSBHYPk"

				var btns []frame.Button

				btns = append(btns, frame.TransactionButton)
				btns = append(btns, frame.PromptButton)
				fmt.Println("btns ", btns)
				returnFrame(w, frmaeId, image, response, btns)
			case 2:
				response, err := frame.ParseFrameAction(button, item, verifiedEthAddress)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				fmt.Println("Response: ", response)
				http.Redirect(w, r, response, http.StatusFound)

			}
			// parseFrameAction(message)

		}
	}
}

type createFrameRequest struct {
	ItemId     string `json:"itemId"`
	ImageUrl   string `json:"imageUrl"`
	Collection string `json:"collection"`
}

func createFrameHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var createFrameReq createFrameRequest
		req := r.Body
		if err := json.NewDecoder(req).Decode(&createFrameReq); err != nil {
			log.Println("error occured decoding request", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		itemId := createFrameReq.ItemId
		imageUrl := createFrameReq.ImageUrl
		collectionAddr := createFrameReq.Collection

		frameId, err := frame.CreateClaimFrame(itemId, imageUrl, collectionAddr, DB)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		baseurl := os.Getenv("BASE_URL")
		url := fmt.Sprintf("%v/frame/%v", baseurl, frameId)

		fmt.Println(url)

		if err := json.NewEncoder(w).Encode(url); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

// serves requests of the form /getframe?itemId
func fetchFrameHandler() http.HandlerFunc {
	type frameObject struct {
		ItemId string `json:"itemId"`
		Url    string `json:"url"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		itemId := r.URL.Query().Get("itemId")
		frameDetail, err := frame.GetFrameByItemId(itemId, DB)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		baseurl := os.Getenv("BASE_URL")
		url := fmt.Sprintf("%v/frame/%v", baseurl, frameDetail.ID)

		response := frameObject{
			ItemId: itemId,
			Url:    url,
		}

		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}

func SetupDatabase() (*gorm.DB, error) {
	// ...
	dsn := os.Getenv("DATABASE_URL")
	fmt.Println("Connecting to database")
	dialector := postgres.Open(dsn)

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		fmt.Println("Error connecting to database")
		return nil, err
	}

	// ...
	if err = db.AutoMigrate(frame.ClaimFrame{}); err != nil {
		log.Println("Error migrating database models")
	}

	return db, nil
}
