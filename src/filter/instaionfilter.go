package filter

type Instationfilter struct {
	Nextfilter Filter
}

func NewInstationfilter() (sf *Instationfilter) {
	sf = new(Instationfilter)
	sf.Nextfilter = nil
	return
}

func (self *Instationfilter) Filt(tmp map[string]string) {
	issuccess := false
	if v, ok := tmp["ref_type"]; ok && v == "0" {
		tmp["rf"] = "0"
		tmp["kw"] = "not provided"
		tmp["secs"] = "not set"
		tmp["rl"] = "no refer"
		tmp["rt"] = "no refer"
		tmp["rl_path"] = "no refer"
		tmp["rl_query"] = "no refer"
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
