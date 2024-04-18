package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/OrlovEvgeny/go-mcache"
	"github.com/ethereum/go-ethereum/common/hexutil"
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
	frame.Cache = mcache.New()

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

func returnFrame(w http.ResponseWriter, frameId, imageUrl, msg string, buttons []frame.Button) {
	w.Header().Set("Content-Type", "text/html")

	ogFrame := frame.ParseFrame(imageUrl, frameId, msg, buttons...)
	fmt.Println("frame ", ogFrame)
	// w.Write()
	fmt.Fprint(w, ogFrame)
}

func validateFrameRequest(frameId, msgData string) (neynar.NeynarFrameValidation, error) {
	neynarApiKey := os.Getenv("NEYNAR_API_KEY")
	neynarClient, err := neynar.NewNeynarClient(neynar.WithApiKey(neynarApiKey))
	if err != nil {
		return neynar.NeynarFrameValidation{}, err
	}

	// frameDetails, err := frame.GetFrameDetails(frameId, DB)
	// if err != nil {
	// 	log.Println(err)
	// 	return err
	// }
	// // framedetails should return dropId
	// imageUrl := frameDetails.ImageUrl
	// drop := frameDetails.ItemId

	type validationBody struct {
		CastReactionContext bool   `json:"cast_reaction_context"`
		FollowContext       bool   `json:"follow_context"`
		MessageBytesInHex   string `json:"message_bytes_in_hex"`
	}
	fmt.Println("msg data", msgData)
	vBody := validationBody{true, false, msgData}
	msgDataBytes, err := json.Marshal(vBody)
	if err != nil {
		log.Println(err)
		return neynar.NeynarFrameValidation{}, err
	}

	action, err := neynarClient.ValidateFrameMessage(msgDataBytes)
	if err != nil {
		log.Printf("error validating frame %v", err)
		return neynar.NeynarFrameValidation{}, err
	}

	if len(action.Interactor.VerifiedAdresses.EthAddresses) == 0 {
		err := neynar.ErrNoVerifications
		log.Println(err)
		return neynar.NeynarFrameValidation{}, err
	}

	return action, nil
}

func frameHandler() http.HandlerFunc {
	type reqBody struct {
		UntrustedData map[string]any    `json:"untrustedData"`
		TrustedData   map[string]string `json:"trustedData"`
	}
	// neynarApiKey := os.Getenv("NEYNAR_API_KEY")
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		fmt.Println("vars", vars)
		frameId, ok := vars["frame"]
		if !ok {
			fmt.Println("id is missing in parameters")
		}
		fmt.Println("frame id", frameId)

		frameDetails, err := frame.GetFrameDetails(frameId, DB)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("an unexpected error occured"))
			return
		}
		imageUrl := frameDetails.ImageUrl
		dropId := frameDetails.DropId

		// fetch drop details
		drop, err := frame.GetDropDetails(dropId, DB)
		if err != nil {
			log.Println(err)
			// w.WriteHeader(http.StatusInternalServerError)
			// w.Write([]byte("an unexpected error occured"))
			// return
		}
		actionRequired := false
		if drop.Criteria != "" {
			actionRequired = true
		}

		fmt.Println("Method", r.Method)

		switch r.Method {
		case http.MethodGet:
			var frameBtn frame.Button

			if actionRequired {
				frameBtn = frame.CheckEligibility
			} else {
				frameBtn = frame.ClaimButton
			}

			if drop.MintPrice != nil || !drop.GasIsCreatorSponsored {
				frame.FrameToExternalClaim(w, imageUrl, *drop.AAContractAddress)
			} else {
				var btns []frame.Button
				btns = append(btns, frameBtn)
				returnFrame(w, frameId, imageUrl, "", btns)
			}

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

			action, err := validateFrameRequest(frameId, msgData)
			if err != nil {
				log.Println(err)
				w.WriteHeader(http.StatusInternalServerError)
			}

			verifiedEthAddress := action.Interactor.VerifiedAdresses.EthAddresses[0]
			// buttonTitle := action.Cast.Frames[0].Buttons[0].Title
			// buttonTitle := "claim"
			uv := r.URL.Query()

			urlAction := uv.Get("action")

			fmt.Println("action - ", urlAction)

			if urlAction == "check-eligibilty" {
				button := frame.Button("check eligibility")

				response, err := frame.ParseFrameAction(button, dropId, verifiedEthAddress, DB)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				// add an artificial delay
				time.Sleep(3 * time.Second)
				var btns []frame.Button
				btns = append(btns, frame.RefreshBotton)
				// drop.FarcasterCriteria
				returnFrame(w, frameId, imageUrl, response, btns)
				return
			} else if urlAction == "refresh" {
				v, ok := frame.Cache.Get(verifiedEthAddress)
				if !ok {
					var btns []frame.Button
					btns = append(btns, frame.RefreshBotton)
					returnFrame(w, frameId, imageUrl, "", btns)
				}
				value := v.(frame.MintPassResponse)
				if value.Valid {
					var btns []frame.Button
					btns = append(btns, frame.ClaimButton)
					returnFrame(w, frameId, imageUrl, "", btns)
				} else {
					// imageUrl = "https://res.cloudinary.com/ludicrousmouse/image/upload/v1712869076/nft_mint_limit_qaozs2.png"

					imageGeneratorUrl := os.Getenv("OG_IMAGE_GENERATOR")
					imageUrl := fmt.Sprintf("%v/frames/opengraph?text=%v&imgUrl=%v", imageGeneratorUrl, value.Message, drop.Image)
					var btns []frame.Button
					btns = append(btns, frame.PromptButton)
					returnFrame(w, frameId, imageUrl, "", btns)
					// } else {
					// 	imageUrl = "https://res.cloudinary.com/ludicrousmouse/image/upload/v1712657526/account_not_elligible_qwieeq.png"
					// 	var btns []frame.Button
					// 	btns = append(btns, frame.PromptButton)
					// 	returnFrame(w, frameId, imageUrl, "", btns)
				}
			} else if urlAction == "claim" {
				button := frame.ClaimButton

				response, err := frame.ParseFrameAction(button, dropId, verifiedEthAddress, DB)
				if err != nil {
					log.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				var btns []frame.Button

				if _, err = hexutil.Decode(response); err != nil {
					fmt.Println(err)
					imageUrl = "https://res.cloudinary.com/ludicrousmouse/image/upload/v1710177216/oops_pfogqm.png"
					btns = append(btns, frame.PromptButton)
					// Todo: use response to generate image
					returnFrame(w, frameId, imageUrl, response, btns)
				}
				btns = append(btns, frame.TransactionButton)
				btns = append(btns, frame.PromptButton)
				returnFrame(w, frameId, imageUrl, response, btns)
			} else if urlAction == "mint" {
			}
		}
	}
}

type createFrameRequest struct {
	DropId     string `json:"dropId"`
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
		dropId := createFrameReq.DropId
		imageUrl := createFrameReq.ImageUrl
		collectionAddr := createFrameReq.Collection

		frameId, err := frame.CreateClaimFrame(dropId, imageUrl, collectionAddr, DB)
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
		frameDetail, err := frame.GetFrameByDropId(itemId, DB)
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

// func parseFarcasterCriteria(farcasterCriteria frame.FarcasterCriteria) {
// 	criterias := strings.Split(farcasterCriteria.CriteriaType, ",")
// 	for _, criteria := range criterias {
// 		switch criteria {
// 		case "farcasterInteractions":
// 			castUrl := farcasterCriteria.CastUrl
// 		}
// 	}
// }

// farcasterInteractions,farcasterFollowing,farcasterChannel
