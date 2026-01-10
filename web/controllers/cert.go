package controllers

import (
	"fmt"
	"ehang.io/nps/lib/file"
	"strings"
)

type CertController struct {
	BaseController
}

// List 证书列表页（兼容路由 /cert/list）
func (s *CertController) List() {
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
		certContent := s.getEscapeString("cert_content")
		keyContent := s.getEscapeString("key_content")

		// 简单验证：检查是否是有效的证书格式
		if !strings.Contains(certContent, "-----BEGIN CERTIFICATE-----") {
			s.AjaxErr("invalid certificate format")
			return
		}
		if !strings.Contains(keyContent, "-----BEGIN PRIVATE KEY-----") && !strings.Contains(keyContent, "-----BEGIN RSA PRIVATE KEY-----") {
			s.AjaxErr("invalid key format")
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
		certContent := s.getEscapeString("cert_content")
		keyContent := s.getEscapeString("key_content")

		// 简单验证：检查是否是有效的证书格式
		if !strings.Contains(certContent, "-----BEGIN CERTIFICATE-----") {
			s.AjaxErr("invalid certificate format")
			return
		}
		if !strings.Contains(keyContent, "-----BEGIN PRIVATE KEY-----") && !strings.Contains(keyContent, "-----BEGIN RSA PRIVATE KEY-----") {
			s.AjaxErr("invalid key format")
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
