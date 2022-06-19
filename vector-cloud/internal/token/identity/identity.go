package identity

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"time"

	"github.com/digital-dream-labs/vector-cloud/internal/log"
	"github.com/digital-dream-labs/vector-cloud/internal/robot"

	jwt "github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc/credentials"
)

const jwtFile = "token.jwt"

// Token provides the methods that clients will care about for authenticating and
// using tokens
type Token interface {
	IssuedAt() time.Time
	RefreshTime() time.Time
	String() string
	UserID() string
}

// Provider is an interface to manage JWT tokens and TLS certs for a single robot
type Provider interface {
	Init() error
	ParseAndStoreToken(token string) (Token, error)
	GetToken() Token
	CertCommonName() string
	TransportCredentials() credentials.TransportCredentials
}

type fileProvider struct {
	credentials    credentials.TransportCredentials
	jwtPath        string
	currentToken   *TokenInfo
	certCommonName string
}

// NewFileProvider creates a new file backed Provider interface implementation
func NewFileProvider(jwtPath, cloudDir string) (*fileProvider, error) {
	if jwtPath == "" {
		jwtPath = DefaultTokenPath
	}
	if cloudDir == "" {
		cloudDir = robot.DefaultCloudDir
	}

	certCommonName, err := robot.CertCommonName(cloudDir)
	if err != nil {
		return nil, err
	}

	credentials, err := getTLSCert(cloudDir)
	if err != nil {
		return nil, err
	}

	return &fileProvider{
		credentials:    credentials,
		jwtPath:        jwtPath,
		certCommonName: certCommonName,
	}, nil
}

func (c *fileProvider) CertCommonName() string {
	return c.certCommonName
}

func (c *fileProvider) TransportCredentials() credentials.TransportCredentials {
	return c.credentials
}

// ParseAndStoreToken parses the given token received from the server and saves it
// to our persistent store
func (c *fileProvider) ParseAndStoreToken(token string) (Token, error) {
	tok, err := c.parseToken(token)
	if err != nil {
		return nil, err
	}
	// everything ok, token is legit
	if err := c.saveToken(token); err != nil {
		return nil, err
	}
	c.currentToken = tok
	logUserID(tok)
	return tokWrapper{tok}, nil
}

// Init triggers the jwt package to initialize its data from disk
func (c *fileProvider) Init() error {
	err := c.init()
	if err != nil {
		if err := robot.WriteFaceErrorCode(851); err != nil {
			log.Println("Couldn't print face error:", err)
		}
	}
	return err
}

// GetToken returns the current loaded token, if there is one. If this returns
// nil, then one should be requested from the server. If not, it might be worth
// checking ShouldRefresh() on the token to see if a new one should be requested
// anyway.
func (c *fileProvider) GetToken() Token {
	if c.currentToken == nil {
		return nil
	}
	return tokWrapper{c.currentToken}
}

func (c *fileProvider) init() error {
	// try to create dir token will live in
	if err := os.Mkdir(c.jwtPath, 0777); err != nil {
		// if this failed, make sure it's because it already exists
		s, err := os.Stat(c.jwtPath)
		if err != nil {
			log.Println("token mkdir + stat error:", err)
			return err
		} else if !s.IsDir() {
			err := fmt.Errorf("token store exists but is not a dir: %s", c.jwtPath)
			log.Println(err)
			return err
		}
	}
	// see if a token already lives on disk
	buf, err := ioutil.ReadFile(c.tokenFile())
	if err == nil {
		tok, err := c.parseToken(string(buf))
		if err != nil {
			os.Remove(c.tokenFile())
			return err
		}

		// TODO DELETE AFTER SEPTEMBER 7TH-ISH
		// delete fake, no-userid token TMS used to generate for testing
		if tok.UserId == "" {
			log.Println("Deleting old test token")
			os.Remove(c.tokenFile())
			return nil
		}

		c.currentToken = tok
		logUserID(tok)
	}
	return nil
}

func (c *fileProvider) tokenFile() string {
	return path.Join(c.jwtPath, jwtFile)
}

func (c *fileProvider) parseToken(token string) (*TokenInfo, error) {
	t, _, err := new(jwt.Parser).ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return nil, err
	}
	tok, err := FromJwtToken(t)
	if err != nil {
		return nil, err
	}
	return tok, nil
}

func (c *fileProvider) saveToken(token string) error {
	if err := os.Mkdir(c.jwtPath, os.ModeDir); err != nil && !os.IsExist(err) {
		return err
	}
	fileName := c.tokenFile()
	tmpFileName := fileName + ".tmp"
	if err := ioutil.WriteFile(tmpFileName, []byte(token), 0777); err != nil {
		return err
	}
	return os.Rename(tmpFileName, fileName)
}

func logUserID(token *TokenInfo) {
	if token == nil {
		return
	}
	if user := token.UserId; user != "" {
		log.Das("profile_id.start", (&log.DasFields{}).SetStrings(user))
	}
}

type tokWrapper struct {
	tok *TokenInfo
}

func (t tokWrapper) RefreshTime() time.Time {
	return t.tok.ExpiresAt.Add(-3 * time.Hour)
}

func (t tokWrapper) String() string {
	return t.tok.Raw
}

func (t tokWrapper) IssuedAt() time.Time {
	return t.tok.IssuedAt
}

func (t tokWrapper) UserID() string {
	return t.tok.UserId
}

var platformTokenPath string
var testTokenPath string

func tokenPath() string {
	if testTokenPath != "" {
		return testTokenPath
	}
	return platformTokenPath
}
