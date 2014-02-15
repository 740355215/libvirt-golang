package libvirt

// #cgo pkg-config: libvirt
// #include <stdlib.h>
// #include <libvirt/libvirt.h>
// #include <libvirt/virterror.h>
import "C"
import (
	"errors"
	"fmt"
	"log"
	"unsafe"
)

// Connection holds a libvirt connection. There are no exported fields.
type Connection struct {
	virConnect C.virConnectPtr
}

// Open creates a new libvirt connection to the Hypervisor. The URIs are
// documented at http://libvirt.org/uri.html.
func Open(uri string) (Connection, error) {
	cUri := C.CString(uri)
	defer C.free(unsafe.Pointer(cUri))

	cConn := C.virConnectOpen(cUri)
	if cConn == nil {
		return Connection{}, fmt.Errorf("libvirt connection to %s failed", uri)
	}

	return Connection{cConn}, nil
}

// OpenReadOnly creates a restricted libvirt connection. The URIs are
// documented at http://libvirt.org/uri.html.
func OpenReadOnly(uri string) (Connection, error) {
	cUri := C.CString(uri)
	defer C.free(unsafe.Pointer(cUri))

	cConn := C.virConnectOpenReadOnly(cUri)
	if cConn == nil {
		return Connection{}, fmt.Errorf("libvirt connection to %s failed", uri)
	}

	return Connection{cConn}, nil
}

// Close closes the connection to the Hypervisor. Connections are reference
// counted; the count is explicitly increased by the initial open (Open,
// OpenAuth, and the like) as well as Ref (not implemented yet); it is also
// temporarily increased by other API that depend on the connection remaining
// alive. The open and every Ref call should have a matching Close, and all
// other references will be released after the corresponding operation
// completes.
// It returns a positive number if at least 1 reference remains on success. The
// returned value should not be assumed to be the total reference count. A
// return of 0 implies no references remain and the connection is closed and
// memory has been freed. It is possible for the last Close to return a
// positive value if some other object still has a temporary reference to the
// connection, but the application should not try to further use a connection
// after the Close that matches the initial open.
func (conn Connection) Close() (int, error) {
	cRet := C.virConnectClose(conn.virConnect)
	ret := int(cRet)

	if ret == -1 {
		return 0, errors.New("failed to close libvirt connection")
	}

	return ret, nil
}

// Version gets the version level of the Hypervisor running.
func (conn Connection) Version() (uint64, error) {
	var cVersion C.ulong
	cRet := C.virConnectGetVersion(conn.virConnect, &cVersion)
	ret := int(cRet)

	if ret == -1 {
		return 0, errors.New("failed to get hypervisor version")
	}

	return uint64(cVersion), nil
}

// LibVersion provides the version of libvirt used by the daemon running on
// the host.
func (conn Connection) LibVersion() (uint64, error) {
	var cVersion C.ulong
	cRet := C.virConnectGetLibVersion(conn.virConnect, &cVersion)
	ret := int(cRet)

	if ret == -1 {
		return 0, errors.New("failed to get libvirt version")
	}

	return uint64(cVersion), nil
}

// IsAlive determines if the connection to the hypervisor is still alive.
// If an error occurs, the function will also return "false" and the error
// message will be written to the log.
func (conn Connection) IsAlive() bool {
	cRet := C.virConnectIsAlive(conn.virConnect)
	ret := int(cRet)

	if ret == 1 {
		return true
	}

	if ret == -1 {
		log.Println("could not check if libvirt connection is alive")
	}

	return false
}

// IsEncrypted determines if the connection to the hypervisor is encrypted.
// If an error occurs, the function will also return "false" and the error
// message will be written to the log.
func (conn Connection) IsEncrypted() bool {
	cRet := C.virConnectIsEncrypted(conn.virConnect)
	ret := int(cRet)

	if ret == 1 {
		return true
	}

	if ret == -1 {
		log.Println("could not check if libvirt connection is encrypted")
	}

	return false
}

// IsSecure determines if the connection to the hypervisor is secure.
// If an error occurs, the function will also return "false" and the error
// message will be written to the log.
func (conn Connection) IsSecure() bool {
	cRet := C.virConnectIsSecure(conn.virConnect)
	ret := int(cRet)

	if ret == 1 {
		return true
	}

	if ret == -1 {
		log.Println("could not check if libvirt connection is secure")
	}

	return false
}

// Capabilities provides capabilities of the hypervisor/driver.
func (conn Connection) Capabilities() (string, error) {
	cCap := C.virConnectGetCapabilities(conn.virConnect)
	if cCap == nil {
		return "", errors.New("failed to get hypervisor capabilities")
	}
	defer C.free(unsafe.Pointer(cCap))

	return C.GoString(cCap), nil
}

// Hostname returns a system hostname on which the hypervisor is running
// (based on the result of the gethostname system call, but possibly expanded
// to a fully-qualified domain name via getaddrinfo). If we are connected to a
// remote system, then this returns the hostname of the remote system.
func (conn Connection) Hostname() (string, error) {
	cHostname := C.virConnectGetHostname(conn.virConnect)
	if cHostname == nil {
		return "", errors.New("failed to get hypervisor hostname")
	}
	defer C.free(unsafe.Pointer(cHostname))

	return C.GoString(cHostname), nil
}

// Sysinfo returns the XML description of the sysinfo details for the host on
// which the hypervisor is running, in the same format as the <sysinfo> element
// of a domain XML. This information is generally available only for
// hypervisors running with root privileges.
func (conn Connection) Sysinfo() (string, error) {
	cSysinfo := C.virConnectGetSysinfo(conn.virConnect, 0)
	if cSysinfo == nil {
		return "", errors.New("failed to get hypervisor sysinfo")
	}
	defer C.free(unsafe.Pointer(cSysinfo))

	return C.GoString(cSysinfo), nil
}

// Type gets the name of the Hypervisor driver used. This is merely the driver
// name; for example, both KVM and QEMU guests are serviced by the driver for
// the qemu:// URI, so a return of "QEMU" does not indicate whether KVM
// acceleration is present. For more details about the hypervisor, use
// Capabilities.
func (conn Connection) Type() (string, error) {
	cType := C.virConnectGetType(conn.virConnect)
	if cType == nil {
		return "", errors.New("failed to get hypervisor type")
	}
	defer C.free(unsafe.Pointer(cType))

	return C.GoString(cType), nil
}

// Uri returns the URI (name) of the hypervisor connection. Normally this is
// the same as or similar to the string passed to the Open/OpenReadOnly call,
// but the driver may make the URI canonical. If uri == "" was passed to Open,
// then the driver will return a non-NULL URI which can be used to connect tos
// the same hypervisor later.
func (conn Connection) Uri() (string, error) {
	cUri := C.virConnectGetURI(conn.virConnect)
	if cUri == nil {
		return "", errors.New("failed to get hypervisor URI")
	}
	defer C.free(unsafe.Pointer(cUri))

	return C.GoString(cUri), nil
}
