package filter

import (
	"net/url"
)

type Outstationfilter struct {
	Nextfilter Filter
}

func NewOutstationfilter() (sf *Outstationfilter) {
	sf = new(Outstationfilter)
	sf.Nextfilter = nil
	return
}

func (self *Outstationfilter) Filt(tmp map[string]string) {
	issuccess := false
	if v, ok := tmp["ref_type"]; ok && v == "1" {
		tmp["rf"] = "3"
		tmp["kw"] = "not provided"
		tmp["secs"] = "not set"
		if v1, ok1 := tmp["ref"]; !ok1 || v1 == "no refer" {
			tmp["rl"] = "no refer"
			tmp["rt"] = "no refer"
			tmp["rl_path"] = "no refer"
			tmp["rl_query"] = "no refer"
		} else {
			tmp["rl"] = v1
			u, err := url.Parse(v1)
			if err != nil {
				tmp["rt"] = "no refer"
				tmp["rl_path"] = "no refer"
				tmp["rl_query"] = "no refer"
			} else {
				tmp["rt"] = u.Host
				tmp["rl_path"] = u.Path
				tmp["rl_query"] = u.RawQuery
			}
		}
		issuccess = true
	}

	if !issuccess && self.Nextfilter != nil {
		self.Nextfilter.Filt(tmp)
	} else if !issuccess && self.Nextfilter == nil {
		tmp["rf"] = "0"
		tmp["kw"] = "not provided"
		tmp["secs"] = "not set"
		tmp["rl"] = "no refer"
		tmp["rt"] = "no refer"
		tmp["rl_path"] = "no refer"
		tmp["rl_query"] = "no refer"
	}
}
