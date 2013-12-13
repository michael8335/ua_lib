package filter

import (
	"common"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"regexp"
	"sort"
	"strings"
)

type Ua struct {
	Id			   string
	Brand          string
	Brand_eng      string
	Ua_name        string
	Ua_pattern     string
	Ua_patterntype string
	Priority       int
}

type Uafilter struct {
	Uainfo     common.SortQueue
	regexobj   map[string]*regexp.Regexp
	Nextfilter Filter
}

func NewUafilter() (sf *Uafilter) {
	sf = new(Uafilter)
	sf.Uainfo = make(common.SortQueue, 0)
	sf.regexobj = make(map[string]*regexp.Regexp)
	sf.Nextfilter = nil
	return
}

func (self *Uafilter) ReadInfo(connstring string) error {
	println(connstring)
	db, err := sql.Open("mysql", connstring)
	if err != nil {
		return err
	}
	defer db.Close()

	sql := "select id, brand, brand_eng, ua_name, ua_pattern, ua_patterntype, priority from conf_ua_info order by brand"
	rows, _ := db.Query(sql)
	for rows.Next() {
		tmp := &Ua{}
		rows.Scan(&tmp.Id, &tmp.Brand, &tmp.Brand_eng, &tmp.Ua_name, &tmp.Ua_pattern, &tmp.Ua_patterntype, &tmp.Priority)
		item := &common.SortItem{tmp, "0"}
		self.Uainfo = append(self.Uainfo, item)

		if tmp.Ua_patterntype == "regex" {
			obj, err := regexp.Compile(tmp.Ua_pattern)
			if err == nil {
				self.regexobj[tmp.Ua_pattern] = obj
			}
		}
	}
	sort.Sort(self.Uainfo)
	return nil
}

func (self *Uafilter) GetUainfo(uastr string) (string, string, string) {
	for _, v := range self.Uainfo {
		ua, ok := v.Value.(*Ua)
		if !ok {
			continue
		}
		if ua.Ua_patterntype == "str" {
			if strings.Contains(uastr, ua.Ua_pattern) {
				return ua.Brand_eng, ua.Ua_name, ua.Id
			}
		} else if ua.Ua_patterntype == "regex" {
			regexobj, ok := self.regexobj[ua.Ua_pattern]
			if ok && regexobj.FindStringIndex(uastr) != nil {
				return ua.Brand_eng, ua.Ua_name, ua.Id
			}
		}
	}
	return "", "", "-1"
}

func (self *Uafilter) Filt(tmp map[string]string) {
	issuccess := false
	v, ok := tmp["ua"]
	if ok {
		if v == "" || v == "-" || v == "not set" {
			tmp["brand_eng"] = "not set"
			tmp["ua_name"] = "not set"
			tmp["ua_conf_id"] = "-1"
		} else {
			brand_eng, ua_name, ua_Id := self.GetUainfo(v)
			if brand_eng != "" && ua_name != "" {
				tmp["brand_eng"] = brand_eng
				tmp["ua_name"] = ua_name
				tmp["ua_conf_id"] = ua_Id
			} else {
				tmp["brand_eng"] = "not set"
				tmp["ua_name"] = "not set"
				tmp["ua_conf_id"] = "-1"
			}
		}
	}

	if !issuccess && self.Nextfilter != nil {
		self.Nextfilter.Filt(tmp)
	}
}
