package filter

import (
	"common"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"regexp"
	"sort"
	"strings"
)

type Browser struct {
	Id					string
	Browser_type        string
	Browser_version     string
	Browser_keyword     string
	Browser_pattern     string
	Browser_patterntype string
	Priority            int
	Keyword_priority    int
}

type Browserfilter struct {
	Browserkeyword common.SortQueue
	Browserinfo    map[string]*common.SortQueue
	regexobj       map[string]*regexp.Regexp
	Nextfilter     Filter
}

func NewBrowserfilter() (sf *Browserfilter) {
	sf = new(Browserfilter)
	sf.Browserkeyword = make(common.SortQueue, 0)
	sf.Browserinfo = make(map[string]*common.SortQueue)
	sf.regexobj = make(map[string]*regexp.Regexp)
	sf.Nextfilter = nil
	return
}

func (self *Browserfilter) ReadInfo(connstring string) error {
	db, err := sql.Open("mysql", connstring)
	if err != nil {
		return err
	}
	defer db.Close()

	sql := "select id, browser_type, browser_version, browser_keyword, browser_pattern, browser_patterntype, priority, keyword_priority from conf_browser_info order by browser_type"
	rows, _ := db.Query(sql)
	for rows.Next() {
		tmp := &Browser{}
		rows.Scan(&tmp.Id, &tmp.Browser_type, &tmp.Browser_version, &tmp.Browser_keyword, &tmp.Browser_pattern, &tmp.Browser_patterntype, &tmp.Priority, &tmp.Keyword_priority)

		if queue, ok := self.Browserinfo[tmp.Browser_keyword]; ok {
			item1 := &common.SortItem{tmp, fmt.Sprintf("%03d", tmp.Priority)}
			*queue = append(*queue, item1)
		} else {
			//Browserkeyword中新增关键字
			item := &common.SortItem{tmp.Browser_keyword, fmt.Sprintf("%03d", tmp.Keyword_priority)}
			self.Browserkeyword = append(self.Browserkeyword, item)

			//创建新的SortQueue
			item1 := &common.SortItem{tmp, fmt.Sprintf("%03d", tmp.Priority)}
			t := make(common.SortQueue, 0)

			t = append(t, item1)
			self.Browserinfo[tmp.Browser_keyword] = &t
		}

		if tmp.Browser_patterntype == "regex" {
			obj, err := regexp.Compile(tmp.Browser_pattern)
			if err == nil {
				self.regexobj[tmp.Browser_pattern] = obj
			}
		}
	}
	sort.Sort(self.Browserkeyword)
	for _, v := range self.Browserinfo {
		sort.Sort(*v)
	}
	return nil
}

func (self *Browserfilter) GetBrowserinfo(uastr string) (string, string, string) {
	for _, v := range self.Browserkeyword {
		oskeyword, ok := v.Value.(string)
		if ok && strings.Contains(uastr, oskeyword) {
			queue, ok := self.Browserinfo[oskeyword]
			if !ok {
				continue
			}
			for _, qv := range *queue {
				browser, ok := qv.Value.(*Browser)
				if !ok {
					continue
				}
				if browser.Browser_patterntype == "str" {
					if strings.Contains(uastr, browser.Browser_pattern) {
						return browser.Browser_type, browser.Browser_version, browser.Id
					}
				} else if browser.Browser_patterntype == "regex" {
					regexobj, ok := self.regexobj[browser.Browser_pattern]
					if ok && regexobj.FindStringIndex(uastr) != nil {
						return browser.Browser_type, browser.Browser_version, browser.Id
					}
				}
			}
		} else {
			continue
		}
	}
	return "", "","-1"
}

func (self *Browserfilter) Filt(tmp map[string]string) {
	issuccess := false
	v, ok := tmp["ua"]
	if ok {
		if v == "" || v == "-" {
			tmp["browser_type"] = "-"
			tmp["browser_version"] = "-"
			tmp["browser_conf_id"] = "-1"
		} else {
			browser_type, browser_version ,browser_id := self.GetBrowserinfo(v)
			if browser_type != "" && browser_version != "" {
				tmp["browser_type"] = browser_type
				tmp["browser_version"] = browser_version
				tmp["browser_conf_id"]=browser_id
			} else {
				tmp["browser_type"] = "-"
				tmp["browser_version"] = "-"
				tmp["browser_conf_id"] = "-1"
			}
		}
	}

	if !issuccess && self.Nextfilter != nil {
		self.Nextfilter.Filt(tmp)
	}
}
