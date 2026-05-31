package translate

import (
	"crypto/md5"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
	"unicode"
)

//go:embed seed_translations.json
var seedData []byte

var translationCache = map[string]string{}
var cacheLoaded = false

func loadCache() {
	if cacheLoaded {
		return
	}
	cacheLoaded = true

	// Load seed translations first (bundled with binary)
	if len(seedData) > 0 {
		var seed map[string]string
		if err := json.Unmarshal(seedData, &seed); err == nil {
			for k, v := range seed {
				translationCache[k] = v
			}
		}
	}

	// Load user cache (takes priority over seed)
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".claude", "ccm-translations.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	json.Unmarshal(data, &translationCache)
}

func saveCache() {
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".claude", "ccm-translations.json")
	data, _ := json.MarshalIndent(translationCache, "", "  ")
	os.WriteFile(path, data, 0644)
}

// SetAPIConfig stores Baidu Translate API credentials so the frontend
// can configure them at runtime.
var apiAppID, apiSecretKey string

func SetAPIConfig(appID, secretKey string) {
	apiAppID = appID
	apiSecretKey = secretKey
}

func IsMostlyChinese(s string) bool {
	cjk := 0
	total := 0
	for _, r := range s {
		if unicode.IsSpace(r) || unicode.IsPunct(r) {
			continue
		}
		total++
		if unicode.Is(unicode.Han, r) {
			cjk++
		}
	}
	if total == 0 {
		return false
	}
	return float64(cjk)/float64(total) > 0.3
}

// TranslateDescription translates an English description to Chinese.
// Returns "" if already Chinese or translation fails.
func TranslateDescription(en string) string {
	if en == "" || IsMostlyChinese(en) {
		return ""
	}
	loadCache()
	if zh, ok := translationCache[en]; ok {
		return zh
	}
	zh := callBaiduAPI(en)
	if zh != "" {
		translationCache[en] = zh
		saveCache()
	}
	return zh
}

func callBaiduAPI(text string) string {
	if apiAppID == "" || apiSecretKey == "" {
		return ""
	}

	salt := fmt.Sprintf("%d", rand.Intn(100000))
	signRaw := apiAppID + text + salt + apiSecretKey
	hash := md5.Sum([]byte(signRaw))
	sign := hex.EncodeToString(hash[:])

	params := url.Values{}
	params.Set("q", text)
	params.Set("from", "auto")
	params.Set("to", "zh")
	params.Set("appid", apiAppID)
	params.Set("salt", salt)
	params.Set("sign", sign)

	client := &http.Client{Timeout: 10 * time.Second}
	fullURL := "https://fanyi-api.baidu.com/api/trans/vip/translate?" + params.Encode()

	resp, err := client.Get(fullURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		ErrorCode   string `json:"error_code"`
		ErrorMsg    string `json:"error_msg"`
		TransResult []struct {
			Src string `json:"src"`
			Dst string `json:"dst"`
		} `json:"trans_result"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return ""
	}
	if result.ErrorCode != "" && result.ErrorCode != "52000" {
		// 52000 = success ; anything else is an error
		return ""
	}
	if len(result.TransResult) > 0 {
		return result.TransResult[0].Dst
	}
	return ""
}

// TranslateBatch translates multiple texts, respecting rate limits.
func TranslateBatch(texts []string) {
	loadCache()
	for _, text := range texts {
		if text == "" || IsMostlyChinese(text) {
			continue
		}
		if _, ok := translationCache[text]; ok {
			continue
		}
		zh := callBaiduAPI(text)
		if zh != "" {
			translationCache[text] = zh
		}
		// Baidu free tier: QPS=1, stay safe with 1.2s interval
		time.Sleep(1200 * time.Millisecond)
	}
	saveCache()
}
