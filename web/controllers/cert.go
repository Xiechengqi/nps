package controllers

import (
	"ehang.io/nps/lib/common"
	"ehang.io/nps/lib/file"
	"fmt"
	"strings"
)

type CertController struct {
	BaseController
}

func resolveCertOrPath(input string) (string, error) {
	if strings.Contains(input, "-----BEGIN CERTIFICATE-----") {
		return input, nil
	}
	if input == "" {
		return "", fmt.Errorf("certificate is empty")
	}
	if !common.FileExists(input) {
		return "", fmt.Errorf("certificate file not found: %s", input)
	}
	b, err := common.ReadAllFromFile(input)
	if err != nil {
		return "", fmt.Errorf("read certificate file error: %w", err)
	}
	content := string(b)
	if !strings.Contains(content, "-----BEGIN CERTIFICATE-----") {
		return "", fmt.Errorf("invalid certificate format")
	}
	return content, nil
}

func resolveKeyOrPath(input string) (string, error) {
	if strings.Contains(input, "-----BEGIN PRIVATE KEY-----") ||
		strings.Contains(input, "-----BEGIN RSA PRIVATE KEY-----") ||
		strings.Contains(input, "-----BEGIN EC PRIVATE KEY-----") ||
		strings.Contains(input, "-----BEGIN ENCRYPTED PRIVATE KEY-----") {
		return input, nil
	}
	if input == "" {
		return "", fmt.Errorf("key is empty")
	}
	if !common.FileExists(input) {
		return "", fmt.Errorf("key file not found: %s", input)
	}
	b, err := common.ReadAllFromFile(input)
	if err != nil {
		return "", fmt.Errorf("read key file error: %w", err)
	}
	content := string(b)
	if !strings.Contains(content, "-----BEGIN PRIVATE KEY-----") &&
		!strings.Contains(content, "-----BEGIN RSA PRIVATE KEY-----") &&
		!strings.Contains(content, "-----BEGIN EC PRIVATE KEY-----") &&
		!strings.Contains(content, "-----BEGIN ENCRYPTED PRIVATE KEY-----") {
		return "", fmt.Errorf("invalid key format")
	}
	return content, nil
}

// List 证书列表页（兼容路由 /cert/list）
func (s *CertController) List() {
	if s.Ctx.Request.Method == "POST" {
		s.GetCertList()
		return
	}
	s.Index()
}

// Index 证书列表页
func (s *CertController) Index() {
	s.Data["menu"] = "cert"
	s.SetInfo("certificate")
	s.display("index/clist")
}

// GetCertList 获取证书列表（表格数据）
func (s *CertController) GetCertList() {
	start, length := s.GetAjaxParams()
	if length == 0 {
		length = 50
	}
	search := s.getEscapeString("search")
	list, cnt := file.GetDb().GetCertList(start, length, search)
	s.AjaxTable(list, cnt, cnt, nil)
}

// Add 添加证书页面/处理
func (s *CertController) Add() {
	if s.Ctx.Request.Method == "GET" {
		s.Data["menu"] = "cert"
		s.SetInfo("add certificate")
		s.display("index/cadd")
	} else {
		certContent, err := resolveCertOrPath(s.getEscapeString("cert_content"))
		if err != nil {
			s.AjaxErr(err.Error())
			return
		}
		keyContent, err := resolveKeyOrPath(s.getEscapeString("key_content"))
		if err != nil {
			s.AjaxErr(err.Error())
			return
		}

		c := &file.DomainCert{
			Name:        s.getEscapeString("name"),
			Domain:      s.getEscapeString("domain"),
			CertContent: certContent,
			KeyContent:  keyContent,
			Remark:      s.getEscapeString("remark"),
		}

		if err := file.GetDb().NewCert(c); err != nil {
			s.AjaxErr(err.Error())
		}
		s.AjaxOk("add success")
	}
}

// Edit 编辑证书页面/处理
func (s *CertController) Edit() {
	id := s.GetIntNoErr("id")
	if s.Ctx.Request.Method == "GET" {
		s.Data["menu"] = "cert"
		if c, err := file.GetDb().GetCertById(id); err != nil {
			s.error()
		} else {
			s.Data["cert"] = c
		}
		s.SetInfo("edit certificate")
		s.display("index/cedit")
	} else {
		certContent, err := resolveCertOrPath(s.getEscapeString("cert_content"))
		if err != nil {
			s.AjaxErr(err.Error())
			return
		}
		keyContent, err := resolveKeyOrPath(s.getEscapeString("key_content"))
		if err != nil {
			s.AjaxErr(err.Error())
			return
		}

		c, err := file.GetDb().GetCertById(id)
		if err != nil {
			s.AjaxErr("certificate not found")
			return
		}

		c.Name = s.getEscapeString("name")
		c.Domain = s.getEscapeString("domain")
		c.CertContent = certContent
		c.KeyContent = keyContent
		c.Remark = s.getEscapeString("remark")

		if err := file.GetDb().UpdateCert(c); err != nil {
			s.AjaxErr(err.Error())
		}
		s.AjaxOk("modified success")
	}
}

// Del 删除证书
func (s *CertController) Del() {
	id := s.GetIntNoErr("id")

	// 检查是否被域名引用
	usedCount := file.GetDb().GetCertUsedCount(id)
	if usedCount > 0 {
		s.AjaxErr(fmt.Sprintf("this certificate is used by %d domain(s), cannot delete", usedCount))
		return
	}

	if err := file.GetDb().DelCert(id); err != nil {
		s.AjaxErr("delete error")
	}
	s.AjaxOk("delete success")
}

// GetCertListForSelect 获取证书列表（下拉框用）
func (s *CertController) GetCertListForSelect() {
	list := file.GetDb().GetCertListForSelect()
	s.Data["json"] = list
	s.ServeJSON()
}
