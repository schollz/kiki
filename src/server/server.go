package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
	"github.com/schollz/kiki/src/database"
	"github.com/schollz/kiki/src/envelope"
	"github.com/schollz/kiki/src/letter"
	"github.com/schollz/kiki/src/logging"
	"github.com/schollz/kiki/src/person"
)

func init() {
	logging.Setup()

}

var (
	DataFolder   = "."
	DatabaseName = "kiki.db"
	Port         = "8003"
	RegionKey    *person.Person
	db           *database.Database
)

func Run() {
	// Setup database
	db = database.Setup(path.Join(DataFolder, DatabaseName))

	// setup kiki instance
	var err error
	RegionKey, err = person.FromPublicPrivateKeys("rbcDfDMIe8qXq4QPtIUtuEylDvlGynx56QgeHUZUZBk=",
		"GQf6ZbBbnVGhiHZ_IqRv0AlfqQh1iofmSyFOcp1ti8Q=") // define region key
	if err != nil {
		panic(err)
	}
	db.Set("AssignedNames", RegionKey.Public(), "Public")

	// Startup server
	r := gin.Default()
	r.HEAD("/", func(c *gin.Context) { // handler for the uptime robot
		c.String(http.StatusOK, "OK")
	})
	r.GET("/identity", handlerIdentity) // handler for generating new identity
	r.POST("/letter", handlerLetter)
	r.POST("/open", handlerOpen)
	r.Run(":" + Port) // listen and serve on 0.0.0.0:8080
}

func respondWithJSON(c *gin.Context, message string, err error) {
	success := err == nil
	if !success {
		message = err.Error()
	}
	c.JSON(http.StatusOK, gin.H{"success": success, "message": message})
}

func handlerIdentity(c *gin.Context) {
	// generate a new person
	p, err := person.New()
	if err != nil {
		panic(err)
	}

	// generate a key for friends
	myfriends, err := person.New()
	if err != nil {
		panic(err)
	}
	myfriendsByte, err := json.Marshal(myfriends)
	// post the key to yourself
	e, err := envelope.SelfAddress(p, "friends-key", string(myfriendsByte))
	if err != nil {
		panic(err)
	}

	// post the envelope
	err = db.AddEnvelope(e)

	// return response
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": err.Error()})
	} else {
		c.JSON(http.StatusOK, p)
	}
}

func handlerOpen(c *gin.Context) {
	respondWithJSON(c, "opened envelopes", handleOpen(c))
}

func handleOpen(c *gin.Context) (err error) {
	// get opener
	file, err := c.FormFile("opener")
	if err != nil {
		return
	}
	data, err := readFormFile(file)
	if err != nil {
		return
	}
	var opener *person.Person
	err = json.Unmarshal(data, &opener)

	// get all the envelopes
	envelopes, err := db.GetEnvelopes()
	if err != nil {
		return
	}
	for _, e := range envelopes {
		ue, err := e.Unseal([]*person.Person{opener, RegionKey})
		if err != nil {
			continue
		}
		fmt.Println(ue.Letter.Content.Kind, ue.Letter.Content.Data)
	}
	return
}

func handlerLetter(c *gin.Context) {
	respondWithJSON(c, "letter added", handleLetter(c))
}

func handleLetter(c *gin.Context) (err error) {
	// get sender
	file, err := c.FormFile("sender")
	if err != nil {
		return
	}
	data, err := readFormFile(file)
	if err != nil {
		return
	}
	var sender *person.Person
	err = json.Unmarshal(data, &sender)

	// get message
	file, err = c.FormFile("message")
	if err != nil {
		return
	}
	data, err = readFormFile(file)
	if err != nil {
		return
	}
	message := string(data)

	// is public
	isPublic := c.PostForm("public") == "yes"

	// get recipients
	recipients := []*person.Person{sender}
	if isPublic {
		recipients = append(recipients, RegionKey)
	}
	recipientsString := c.PostForm("recipients")
	if recipientsString != "" {
		var recipientPublicKeys []string
		err = json.Unmarshal([]byte(recipientsString), &recipientPublicKeys)
		if err != nil {
			return
		}
		for _, recipientPublicKeyString := range recipientPublicKeys {
			otherRecipient, err := person.FromPublicKey(recipientPublicKeyString)
			if err != nil {
				logging.Log.Infof("not a valid public key: '%s'", recipientPublicKeyString)
				continue
			}
			recipients = append(recipients, otherRecipient)
		}
	}

	// make letter
	logging.Log.Info("writing letter")
	l, err := letter.New("post", message, sender.Public())
	if err != nil {
		return
	}

	// seal envelope
	logging.Log.Info("sealing envelope")
	e, err := envelope.New(l, sender, recipients)
	if err != nil {
		return
	}

	// add envelope to database
	logging.Log.Info("putting in carrier")
	err = db.AddEnvelope(e)
	if err != nil {
		return
	}
	return
}

func readFormFile(file *multipart.FileHeader) (data []byte, err error) {
	src, err := file.Open()
	if err != nil {
		return
	}
	defer src.Close()
	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, src)
	data = buf.Bytes()
	return
}
