package smms

import (
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
)

// 懒得写mock server了，想跑单侧，填入自己的账号密码
const (
	username = "xxx"
	password = "xxx"
)

var (
	testFilename string
)

func init() {
	randSrc := rand.NewSource(time.Now().UnixNano())
	r := rand.New(randSrc)
	testFilename = fmt.Sprintf("smms-go_UT_%v.jpg", r.Uint32())
}

func TestInvalidAccount(t *testing.T) {
	rsp, err := Token("Y*R#*HFVD", "HFUI&*#*+?VD")
	if rsp != "" || err == nil {
		t.Fatal("expected a invalid token request")
	}
}

func hasFilename(imgs []ImageRsp, filename string) bool {
	for i := range imgs {
		if imgs[i].Filename == filename {
			return true
		}
	}
	return false
}

func Test(t *testing.T) {
	c, err := New(username, password)
	if err != nil {
		t.Fatalf("expecated a valid account")
	}

	profileRsp, err := c.Profile()
	if err != nil || profileRsp.Username != username {
		t.Fatalf("expecated a valid profile request: %s", spew.Sdump(profileRsp, err))
	}

	img, _ := os.Open("./test.jpg")
	defer img.Close()

	uploadRsp, err := c.Upload(testFilename, img)
	if err != nil || uploadRsp.Filename != testFilename {
		t.Fatalf("expecated a valid upload request: %s", spew.Sdump(uploadRsp, err))
	}
	imgHash := uploadRsp.Hash

	historyRsp, err := c.History()
	if err != nil || !hasFilename(historyRsp, testFilename) {
		t.Fatalf("expecated a valid history request: %s", spew.Sdump(historyRsp, err))
	}

	err = c.Clear()
	if err != nil {
		t.Fatalf("expecated a valid clear request: %s", spew.Sdump(err))
	}

	uploadHistoryRsp, err := c.UploadHistory()
	if err != nil || hasFilename(uploadHistoryRsp, testFilename)  {
		t.Fatalf("expecated a valid upload_history request: %s", spew.Sdump(uploadHistoryRsp, err))
	}

	err = c.Delete(imgHash)
	if err != nil {
		t.Fatalf("expecated a valid delete request: %s", spew.Sdump(err))
	}

	err = c.Delete("invalid-hash-value")
	if err == nil {
		t.Fatalf("expecated a invalid delete request")
	}
}