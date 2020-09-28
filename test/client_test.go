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
	"net/url"
	"strconv"
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
	user   User
	token  string
	ip     string
	bucket string
}

func (suite *ClientTestSuite) SetupSuite() {
	suite.ip = "http://118.31.64.83:8080"
	suite.bucket = "buckettest1"
	suite.user = User{Username: "test1", Password: "test1"}
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
	user := suite.user
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
	request, err := http.NewRequest("GET", suite.ip+"/createbucket/"+suite.bucket, nil)
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
	request, err := http.NewRequest("POST", suite.ip+"/upload/"+suite.bucket+"/test1", body)
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
	request, err := http.NewRequest("GET", suite.ip+"/download/"+suite.bucket+"/test1multipart", nil)
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

type UploadID struct {
	ID string `json:"uploadID"`
}

func (suite *ClientTestSuite) TestMultipartUploadComplete() {
	token := suite.LoginWithJwt()
	uploadId := suite.CreateMultipartUpload(token)
	num, parts := suite.SpliteObject()
	var etags = make([]string, num)
	for i := 0; i < num; i++ {
		client := http.Client{}
		apiUrl := suite.ip + "/uploads/upload/" + suite.bucket + "/" + "test1multipart"
		u, err := url.ParseRequestURI(apiUrl)
		require.NoError(suite.T(), err)
		data := url.Values{}
		data.Set("PartNumber", strconv.Itoa(i))
		data.Set("UploadId", uploadId)
		u.RawQuery = data.Encode()
		body := bytes.NewReader(parts[i])
		request, err := http.NewRequest("POST", u.String(), body)
		require.NoError(suite.T(), err)
		checkSum := md5.New()
		checkSum.Write(parts[i])
		hash := base64.StdEncoding.EncodeToString(checkSum.Sum(nil))
		fmt.Println(hash)
		request.Header.Add("Content-MD5", hash)
		request.Header.Add("Authorization", "Bearer "+token)
		request.Header.Add("c-meta-hello", "hello meta")
		request.Header.Add("C-Acl", "PUBLIC")
		rep, err := client.Do(request)
		if assert.NoError(suite.T(), err) {
			etag := rep.Header.Get("ETag")
			etags = append(etags, etag)
		}
	}

	var multipart CompleteMultipart
	for i := 0; i < num; i++ {
		partId := strconv.Itoa(i)
		etag := etags[i]
		part := Part{
			PartID: partId,
			ETag:   etag,
		}
		multipart.Parts = append(multipart.Parts, part)
	}
	suite.CompleteMultipartUpload(uploadId, suite.bucket, "test1multipart", multipart, token)
}

func (suite *ClientTestSuite) CreateMultipartUpload(token string) string {
	client := http.Client{}
	request, err := http.NewRequest("POST", suite.ip+"/uploads/create/"+suite.bucket+"/test1multipart", nil)
	require.NoError(suite.T(), err)
	request.Header.Add("Authorization", "Bearer "+token)
	rep, err := client.Do(request)
	require.NoError(suite.T(), err)
	body, err := ioutil.ReadAll(rep.Body)
	require.NoError(suite.T(), err)
	var uploadID UploadID
	err = json.Unmarshal(body, &uploadID)
	require.NoError(suite.T(), err)
	return uploadID.ID
}

func (suite *ClientTestSuite) SpliteObject() (int, [][]byte) {
	content, err := ioutil.ReadFile("../testdata/flowers.png")
	require.NoError(suite.T(), err)
	size := 100000
	length := len(content)
	n := 0
	if length%size > 0 {
		n = length/size + 1
	} else {
		n = length / size
	}
	var parts = make([][]byte, n)
	num := 0
	for num < n {
		if num == n-1 {
			parts[num] = append(parts[num], content[num*size:]...)
		} else {
			parts[num] = append(parts[num], content[num*size:(num+1)*size]...)
		}
		num++
	}
	return num, parts
}

type Part struct {
	PartID string `json:"PartID"`
	ETag   string `json:"ETag"`
}

type CompleteMultipart struct {
	Parts []Part
}

func (suite *ClientTestSuite) CompleteMultipartUpload(uploadID, bucket, object string, multipart CompleteMultipart, token string) {
	data, err := json.Marshal(&multipart)
	if err != nil {
		fmt.Println(err)
	}
	body := bytes.NewReader(data)
	client := http.Client{}
	apiUrl := suite.ip + "/uploads/complete/" + bucket + "/" + object
	u, err := url.ParseRequestURI(apiUrl)
	require.NoError(suite.T(), err)
	v := url.Values{}
	v.Set("UploadId", uploadID)
	u.RawQuery = v.Encode()
	request, err := http.NewRequest("POST", u.String(), body)
	require.NoError(suite.T(), err)
	request.Header.Add("Authorization", "Bearer "+token)
	rep, err := client.Do(request)
	require.NoError(suite.T(), err)
	fmt.Printf("%v", rep.StatusCode)
}

func (suite *ClientTestSuite) AbortMultipartUpload(uploadID, bucket, object, token string) {
	client := http.Client{}
	apiUrl := suite.ip + "/uploads/abort/" + bucket + "/" + object
	u, err := url.ParseRequestURI(apiUrl)
	require.NoError(suite.T(), err)
	v := url.Values{}
	v.Set("UploadId", uploadID)
	u.RawQuery = v.Encode()
	request, err := http.NewRequest("POST", u.String(), nil)
	require.NoError(suite.T(), err)
	request.Header.Add("Authorization", "Bearer "+token)
	rep, err := client.Do(request)
	require.NoError(suite.T(), err)
	fmt.Printf("%v", rep.StatusCode)
}

func (suite *ClientTestSuite) TestMultipartUploadAbort() {
	token := suite.LoginWithJwt()
	uploadId := suite.CreateMultipartUpload(token)
	num, parts := suite.SpliteObject()
	var etags = make([]string, num)
	for i := 0; i < num; i++ {
		client := http.Client{}
		apiUrl := suite.ip + "/uploads/upload/" + suite.bucket + "/" + "test1multipart"
		u, err := url.ParseRequestURI(apiUrl)
		require.NoError(suite.T(), err)
		data := url.Values{}
		data.Set("PartNumber", strconv.Itoa(i))
		data.Set("UploadId", uploadId)
		u.RawQuery = data.Encode()
		body := bytes.NewReader(parts[i])
		request, err := http.NewRequest("POST", u.String(), body)
		require.NoError(suite.T(), err)
		checkSum := md5.New()
		checkSum.Write(parts[i])
		hash := base64.StdEncoding.EncodeToString(checkSum.Sum(nil))
		fmt.Println(hash)
		request.Header.Add("Content-MD5", hash)
		request.Header.Add("Authorization", "Bearer "+token)
		request.Header.Add("c-meta-hello", "hello meta")
		request.Header.Add("C-Acl", "PUBLIC")
		rep, err := client.Do(request)
		if assert.NoError(suite.T(), err) {
			etag := rep.Header.Get("ETag")
			etags = append(etags, etag)
		}
	}

	suite.AbortMultipartUpload(uploadId, suite.bucket, "test1multipart", token)
}
