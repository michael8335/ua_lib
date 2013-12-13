package filter

import (
	"common"
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"net/url"
	"sort"
	"strings"
)

type Sns struct {
	Sns_name     string
	Sns_info     string
	Sns_domain   string
	Sns_priority int
}

type Snsfilter struct {
	Snsinfo    common.SortQueue
	Nextfilter Filter
}

func NewSnsfilter() (sf *Snsfilter) {
	sf = new(Snsfilter)
	sf.Snsinfo = make(common.SortQueue, 0)
	sf.Nextfilter = nil
	return
}

func (self *Snsfilter) ReadInfo(connstring string) error {
	db, err := sql.Open("mysql", connstring)
	if err != nil {
		return err
	}
	defer db.Close()

	sql := "select sns_name, sns_info, sns_domain from conf_sns_info order by sns_name"
	rows, _ := db.Query(sql)
	for rows.Next() {
		tmp := &Sns{}
		rows.Scan(&tmp.Sns_name, &tmp.Sns_info, &tmp.Sns_domain)
		item := &common.SortItem{tmp, "0"}
		self.Snsinfo = append(self.Snsinfo, item)
	}
	sort.Sort(self.Snsinfo)
	return nil
}

func (self *Snsfilter) GetSnsinfo(urls string) (string, error) {
	u, err := url.Parse(urls)
	if err != nil {
		return "", err
	}
	for _, v := range self.Snsinfo {
		sns, ok := v.Value.(*Sns)
		if ok && strings.Contains(u.Host, sns.Sns_domain) {
			return sns.Sns_name, nil
		}
	}
	return "", nil
}

func (self *Snsfilter) Filt(tmp map[string]string) {
	issuccess := false
	v, ok := tmp["ref"]
	if ok {
		if v == "" || v == "-" {
			tmp["rl"] = "-"
			tmp["rl_path"] = "-"
			tmp["rl_query"] = "-"
		} else {
			result, err := self.GetSnsinfo(v)
			if err == nil && result != "" {
				tmp["rf"] = "2"
				tmp["rt"] = result
				tmp["kw"] = "-"
				tmp["secs"] = "-"
				u, err := url.Parse(tmp["ref"])
				if err != nil {
					tmp["rl"] = tmp["ref"]
					tmp["rl_path"] = u.Path
					tmp["rl_query"] = u.RawQuery
				} else {
					tmp["rl"] = "-"
					tmp["rl_path"] = "-"
					tmp["rl_query"] = "-"
				}
			}
			issuccess = true
		}
	}

	if !issuccess && self.Nextfilter != nil {
		self.Nextfilter.Filt(tmp)
	} else if !issuccess && self.Nextfilter == nil {
		tmp["rf"] = "10"
		tmp["rt"] = "-"
		tmp["rl"] = "-"
		tmp["rl_path"] = "-"
		tmp["rl_query"] = "-"
		tmp["kw"] = "-"
		tmp["secs"] = "-"
	}
}
