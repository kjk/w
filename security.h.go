// THIS FILE WAS AUTO-GENERATED BY https://github.com/kjk/w/cmd/gengo
package w

import (
	"unsafe"
)

const (
	SecurityAnonymous      = 0
	SecurityIdentification = 1
	SecurityImpersonation  = 2
	SecurityDelegation     = 3
)

type ACL struct {
	AclRevision uint8
	Sbz1        uint8
	AclSize     uint16
	AceCount    uint16
	Sbz2        uint16
}

const (
	SE_OWNER_DEFAULTED       = 0x0001
	SE_GROUP_DEFAULTED       = 0x0002
	SE_DACL_PRESENT          = 0x0004
	SE_DACL_DEFAULTED        = 0x0008
	SE_SACL_PRESENT          = 0x0010
	SE_SACL_DEFAULTED        = 0x0020
	SE_DACL_AUTO_INHERIT_REQ = 0x0100
	SE_SACL_AUTO_INHERIT_REQ = 0x0200
	SE_DACL_AUTO_INHERITED   = 0x0400
	SE_SACL_AUTO_INHERITED   = 0x0800
	SE_DACL_PROTECTED        = 0x1000
	SE_SACL_PROTECTED        = 0x2000
	SE_RM_CONTROL_VALID      = 0x4000
	SE_SELF_RELATIVE         = 0x8000
)

type SECURITY_DESCRIPTOR struct {
	Revision uint8
	Sbz1     uint8
	Control  uint16
	Owner    unsafe.Pointer
	Group    unsafe.Pointer
	Sacl     *ACL
	Dacl     *ACL
}

type SID_IDENTIFIER_AUTHORITY struct {
	Value [6]uint8
}

type SECURITY_ATTRIBUTES struct {
	NLength              uint32
	LpSecurityDescriptor *SECURITY_DESCRIPTOR
	BInheritHandle       int32
}
