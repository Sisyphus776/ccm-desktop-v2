package translate

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

var translationCache = map[string]string{}
var cacheLoaded = false
var lastAPICall time.Time

func loadCache() {
	if cacheLoaded {
		return
	}
	cacheLoaded = true
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

func TranslateDescription(en string) string {
	if en == "" || IsMostlyChinese(en) {
		return ""
	}
	loadCache()
	if zh, ok := translationCache[en]; ok {
		return zh
	}
	zh := callYoudaoAPI(en)
	if zh == "" || !IsMostlyChinese(zh) {
		zh = miniTranslate(en)
	}
	if zh != "" {
		translationCache[en] = zh
		saveCache()
	}
	return zh
}

func callYoudaoAPI(text string) string {
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		if elapsed := time.Since(lastAPICall); elapsed < 500*time.Millisecond {
			time.Sleep(500*time.Millisecond - elapsed)
		}
		lastAPICall = time.Now()
		client := &http.Client{Timeout: 15 * time.Second}
		form := url.Values{}
		form.Set("q", text)
		form.Set("from", "Auto")
		form.Set("to", "Auto")
		req, _ := http.NewRequest("POST", "https://aidemo.youdao.com/trans", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Set("User-Agent", "Mozilla/5.0")
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		var result struct {
			ErrorCode   interface{} `json:"errorCode"`
			Translation []string    `json:"translation"`
		}
		if err := json.Unmarshal(body, &result); err != nil {
			continue
		}
		if len(result.Translation) > 0 {
			if result.ErrorCode == nil {
				return result.Translation[0]
			}
			if fmt.Sprintf("%v", result.ErrorCode) == "0" {
				return result.Translation[0]
			}
		}
	}
	return ""
}

func miniTranslate(en string) string {
	r := en
	changed := false
	for _, p := range miniDict {
		if strings.Contains(r, p.en) {
			r = strings.ReplaceAll(r, p.en, p.zh)
			changed = true
		}
	}
	if !changed {
		return ""
	}
	return r
}

var miniDict = []struct{ en, zh string }{
	{"Use this skill whenever the user wants to", "当用户需要以下操作时使用此 Skill："},
	{"Use this skill any time a", "当涉及"},
	{"Use this skill whenever", "当用户需要时使用此 Skill："},
	{"use when you need to", "当您需要以下操作时使用："},
	{"use when the user needs to", "当用户需要以下操作时使用："},
	{"use this skill when", "在以下场景使用此 Skill："},
	{"Use when about to claim work is complete", "在确认工作完成、声称修复完毕时"},
	{"Use when creating new skills", "在创建新 Skill 时"},
	{"Use when encountering any bug", "在遇到任何 bug 时"},
	{"Use when executing implementation plans", "在执行实施计划时"},
	{"Use when implementing any feature", "在实现任何功能时"},
	{"Use when receiving code review feedback", "在收到代码审查反馈时"},
	{"Use when starting any conversation", "在开始任何对话时"},
	{"Use when starting feature work", "在开始功能开发时"},
	{"Use when you have a spec or requirements", "在有需求规格时"},
	{"Use when", "使用时机："},
	{"before committing or creating PRs", "在提交或创建 PR 之前"},
	{"before writing implementation code", "在编写实现代码之前"},
	{"before touching code", "在修改代码之前"},
	{"before proposing fixes", "在提出修复方案之前"},
	{"Product Requirements Document", "产品需求文档"},
	{"browser automation", "浏览器自动化"},
	{"Word documents", "Word 文档"},
	{"spreadsheet file", "表格文件"},
	{"frontend design", "前端设计"},
	{"code review feedback", "代码审查反馈"},
	{"implementation plans", "实施计划"},
	{"feature work", "功能开发"},
	{"skill creation", "Skill 创建"},
	{"PRD", "产品需求文档"},
	{"PDF files", "PDF 文件"},
	{"slide deck", "幻灯片"},
	{"presentation", "演示文稿"},
	{"create or edit", "创建或编辑"},
	{"creating or editing", "创建或编辑"},
	{"interact with websites", "与网站交互"},
	{"extract content from", "提取内容"},
	{"convert between", "格式转换"},
	{"verifying skills work", "验证 Skill 正常工作"},
	{"editing existing skills", "编辑已有 Skill"},
	{"test failure", "测试失败"},
	{"unexpected behavior", "意外行为"},
	{"independent tasks", "独立任务"},
	{"current session", "当前会话"},
	{"from current workspace", "从当前工作区"},
	{"establishes how to find", "建立如何查找"},
	{"multi-step task", "多步骤任务"},
	{"implementation", "实现"},
	{"requirements", "需求"},
	{"conversation", "对话"},
	{"isolation", "隔离"},
	{"bugfix", "Bug 修复"},
	{"bug or", "Bug 或"},
	{"bug", "Bug"},
	{"SKILL.md", "SKILL.md"},
}
