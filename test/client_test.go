package test

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"net/http"
	"testing"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Result struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

type ClientTestSuite struct {
	suite.Suite
	user  User
	token string
	ip    string
}

func (suite *ClientTestSuite) SetupSuite() {
	suite.ip = "http://118.31.64.83:8080"
}

func (suite *ClientTestSuite) SetupTest() {

}

func (suite *ClientTestSuite) TearDownTest() {

}

func (suite *ClientTestSuite) TearDownSuite() {

}

func TestClientTestSuite(t *testing.T) {
	suite.Run(t, new(ClientTestSuite))
}

func (suite *ClientTestSuite) TestRegisterWithJwt() {
	user := User{Username: "test1", Password: "test1"}
	data, err := json.Marshal(user)
	require.NoError(suite.T(), err)
	client := http.Client{}
	body := bytes.NewReader(data)
	request, err := http.NewRequest("POST", suite.ip+"/register", body)
	require.NoError(suite.T(), err)
	request.Header.Add("Content-Type", "application/json")
	rep, err := client.Do(request)
	if assert.NoError(suite.T(), err) {
		res, _ := ioutil.ReadAll(rep.Body)
		fmt.Printf("%s", res)
		_ = rep.Body.Close()
	}
}

func (suite *ClientTestSuite) LoginWithJwt() string {
	user := User{Username: "test1", Password: "test1"}
	data, err := json.Marshal(user)
	require.NoError(suite.T(), err)
	client := http.Client{}
	body := bytes.NewReader(data)
	request, err := http.NewRequest("POST", suite.ip+"/login", body)
	require.NoError(suite.T(), err)
	rep, err := client.Do(request)
	require.NoError(suite.T(), err)
	res, _ := ioutil.ReadAll(rep.Body)
	var result Result
	err = json.Unmarshal(res, &result)
	if assert.NoError(suite.T(), err) {
		//token
		suite.token = result.Data
		return suite.token
	}
	return ""
}

func (suite *ClientTestSuite) TestCreateBucket() {
	token := suite.LoginWithJwt()
	client := http.Client{}
	request, err := http.NewRequest("GET", suite.ip+"/createbucket/"+"buckettest1", nil)
	require.NoError(suite.T(), err)
	request.Header.Add("C-Acl", "PUBLIC_READ")
	request.Header.Add("Authorization", "Bearer "+token)
	rep, err := client.Do(request)
	require.NoError(suite.T(), err)
	res, err := ioutil.ReadAll(rep.Body)
	if assert.NoError(suite.T(), err) {
		fmt.Printf("%v", string(res))
		_ = rep.Body.Close()
	}
}

func (suite *ClientTestSuite) TestUpload() {
	token := suite.LoginWithJwt()
	content, err := ioutil.ReadFile("../testdata/flowers.png")
	require.NoError(suite.T(), err)
	client := http.Client{}
	body := bytes.NewReader(content)
	request, err := http.NewRequest("POST", suite.ip+"/upload/buckettest1/test1", body)
	if err != nil {
		fmt.Println(err)
	}
	checkSum := md5.New()
	checkSum.Write(content)
	hash := base64.StdEncoding.EncodeToString(checkSum.Sum(nil))
	request.Header.Add("Content-MD5", hash)
	request.Header.Add("c-meta-hello", "hello meta")
	request.Header.Add("Authorization", "Bearer "+token)
	request.Header.Add("C-Acl", "PUBLIC")
	rep, err := client.Do(request)
	require.NoError(suite.T(), err)
	res, err := ioutil.ReadAll(rep.Body)
	if assert.NoError(suite.T(), err) {
		fmt.Println(rep.StatusCode)
		fmt.Printf("%v", string(res))
		_ = rep.Body.Close()
	}
}

func (suite *ClientTestSuite) TestDownload() {
	token := suite.LoginWithJwt()
	client := http.Client{}
	request, err := http.NewRequest("GET", suite.ip+"/download/buckettest1/test1", nil)
	require.NoError(suite.T(), err)
	request.Header.Add("Authorization", "Bearer "+token)
	rep, err := client.Do(request)
	if assert.NoError(suite.T(), err) {
		data, err := ioutil.ReadAll(rep.Body)
		if assert.NoError(suite.T(), err) {
			err = ioutil.WriteFile("../testdata/flowers_download.png", data, 0777)
			require.NoError(suite.T(), err)
		}
	}
}
