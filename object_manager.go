package ibclient

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
)

type IBObjectManager interface {
	CreateNetworkView(name string) (*NetworkView, error)
	CreateDefaultNetviews(globalNetview string, localNetview string) (globalNetviewRef string, localNetviewRef string, err error)
	CreateNetwork(netview string, cidr string, name string) (*Network, error)
	CreateNetworkContainer(netview string, cidr string) (*NetworkContainer, error)
	GetNetworkView(name string) (*NetworkView, error)
	GetNetwork(netview string, cidr string, ea EA) (*Network, error)
	GetNetworkContainer(netview string, cidr string) (*NetworkContainer, error)
	AllocateIP(netview string, cidr string, ipAddr string, macAddress string, name string, vmID string, vmName string) (*FixedAddress, error)
	AllocateNetwork(netview string, cidr string, prefixLen uint, name string) (network *Network, err error)
	UpdateFixedAddress(fixedAddrRef string, matchclient string, macAddress string, vmID string, vmName string) (*FixedAddress, error)
	GetFixedAddress(netview string, cidr string, ipAddr string, macAddr string) (*FixedAddress, error)
	GetFixedAddressByRef(ref string) (*FixedAddress, error)
	DeleteFixedAddress(ref string) (string, error)
	ReleaseIP(netview string, cidr string, ipAddr string, macAddr string) (string, error)
	DeleteNetwork(ref string, netview string) (string, error)
	GetEADefinition(name string) (*EADefinition, error)
	CreateEADefinition(eadef EADefinition) (*EADefinition, error)
	UpdateNetworkViewEA(ref string, addEA EA, removeEA EA) error
	CreateHostRecord(enabledns bool, recordName string, netview string, dnsview string, cidr string, ipAddr string, macAddress string, vmID string, vmName string) (*HostRecord, error)
	GetHostRecordByRef(ref string) (*HostRecord, error)
	GetHostRecord(recordName string, netview string, cidr string, ipAddr string) (*HostRecord, error)
	GetIpAddressFromHostRecord(host HostRecord) (string, error)
	UpdateHostRecord(hostRref string, ipAddr string, macAddress string, vmID string, vmName string) (string, error)
	DeleteHostRecord(ref string) (string, error)
	CreateARecord(netview string, dnsview string, recordname string, cidr string, ipAddr string, vmID string, vmName string) (*RecordA, error)
	GetARecordByRef(ref string) (*RecordA, error)
	DeleteARecord(ref string) (string, error)
	CreateCNAMERecord(canonical string, recordname string, dnsview string) (*RecordCNAME, error)
	GetCNAMERecordByRef(ref string) (*RecordA, error)
	DeleteCNAMERecord(ref string) (string, error)
	CreatePTRRecord(netview string, dnsview string, recordname string, cidr string, ipAddr string, vmID string, vmName string) (*RecordPTR, error)
	GetPTRRecordByRef(ref string) (*RecordPTR, error)
	DeletePTRRecord(ref string) (string, error)
}

type ObjectManager struct {
	connector IBConnector
	cmpType   string
	tenantID  string
	// If OmitCloudAttrs is true no extra attributes for cloud are set
	OmitCloudAttrs bool
}

func NewObjectManager(connector IBConnector, cmpType string, tenantID string) *ObjectManager {
	objMgr := new(ObjectManager)

	objMgr.connector = connector
	objMgr.cmpType = cmpType
	objMgr.tenantID = tenantID
	objMgr.OmitCloudAttrs = true

	return objMgr
}

func NewLocalObjectManager(connector IBConnector) *ObjectManager {
	return &ObjectManager{
		connector:      connector,
		OmitCloudAttrs: true,
	}
}

func (objMgr *ObjectManager) getBasicEA(cloudAPIOwned Bool) EA {
	ea := make(EA)
	if !objMgr.OmitCloudAttrs {
		ea["Cloud API Owned"] = cloudAPIOwned
		ea["CMP Type"] = objMgr.cmpType
		ea["Tenant ID"] = objMgr.tenantID
	}
	return ea
}

func (objMgr *ObjectManager) getBasicVMEA(cloudAPIOwned Bool, vmID, vmName string) EA {
	ea := objMgr.getBasicEA(cloudAPIOwned)
	if !objMgr.OmitCloudAttrs {
		if vmID != "" {
			ea["VM ID"] = vmID
		}

		if vmName != "" {
			ea["VM Name"] = vmName
		}
	}
	return ea
}

func (objMgr *ObjectManager) CreateNetworkView(name string) (*NetworkView, error) {
	networkView := NewNetworkView(NetworkView{
		Name: name,
		Ea:   objMgr.getBasicEA(false)})

	ref, err := objMgr.connector.CreateObject(networkView)
	networkView.Ref = ref

	return networkView, err
}

func (objMgr *ObjectManager) makeNetworkView(netviewName string) (netviewRef string, err error) {
	var netviewObj *NetworkView
	if netviewObj, err = objMgr.GetNetworkView(netviewName); err != nil {
		return
	}
	if netviewObj == nil {
		if netviewObj, err = objMgr.CreateNetworkView(netviewName); err != nil {
			return
		}
	}

	netviewRef = netviewObj.Ref

	return
}

func (objMgr *ObjectManager) CreateDefaultNetviews(globalNetview string, localNetview string) (globalNetviewRef string, localNetviewRef string, err error) {
	if globalNetviewRef, err = objMgr.makeNetworkView(globalNetview); err != nil {
		return
	}

	if localNetviewRef, err = objMgr.makeNetworkView(localNetview); err != nil {
		return
	}

	return
}

func (objMgr *ObjectManager) CreateNetwork(netview string, cidr string, name string) (*Network, error) {
	network := NewNetwork(Network{
		NetviewName: netview,
		Cidr:        cidr,
		Ea:          objMgr.getBasicEA(true)})

	if name != "" {
		network.Ea["Network Name"] = name
	}
	ref, err := objMgr.connector.CreateObject(network)
	if err != nil {
		return nil, err
	}
	network.Ref = ref

	return network, err
}

func (objMgr *ObjectManager) CreateNetworkContainer(netview string, cidr string) (*NetworkContainer, error) {
	container := NewNetworkContainer(NetworkContainer{
		NetviewName: netview,
		Cidr:        cidr,
		Ea:          objMgr.getBasicEA(true)})

	ref, err := objMgr.connector.CreateObject(container)
	container.Ref = ref

	return container, err
}

func (objMgr *ObjectManager) GetNetworkView(name string) (*NetworkView, error) {
	var res []NetworkView

	netview := NewNetworkView(NetworkView{Name: name})

	err := objMgr.connector.GetObject(netview, "", &res)

	if err != nil || res == nil || len(res) == 0 {
		return nil, err
	}

	return &res[0], nil
}

func (objMgr *ObjectManager) UpdateNetworkViewEA(ref string, addEA EA, removeEA EA) error {
	var res NetworkView

	nv := NetworkView{}
	nv.returnFields = []string{"extattrs"}
	err := objMgr.connector.GetObject(&nv, ref, &res)

	if err != nil {
		return err
	}

	for k, v := range addEA {
		res.Ea[k] = v
	}

	for k := range removeEA {
		_, ok := res.Ea[k]
		if ok {
			delete(res.Ea, k)
		}
	}

	_, err = objMgr.connector.UpdateObject(&res, ref)
	return err
}

func BuildNetworkViewFromRef(ref string) *NetworkView {
	// networkview/ZG5zLm5ldHdvcmtfdmlldyQyMw:global_view/false
	r := regexp.MustCompile(`networkview/\w+:([^/]+)/\w+`)
	m := r.FindStringSubmatch(ref)

	if m == nil {
		return nil
	}

	return &NetworkView{
		Ref:  ref,
		Name: m[1],
	}
}

func BuildNetworkFromRef(ref string) *Network {
	// network/ZG5zLm5ldHdvcmskODkuMC4wLjAvMjQvMjU:89.0.0.0/24/global_view
	r := regexp.MustCompile(`network/\w+:(\d+\.\d+\.\d+\.\d+/\d+)/(.+)`)
	m := r.FindStringSubmatch(ref)

	if m == nil {
		return nil
	}

	return &Network{
		Ref:         ref,
		NetviewName: m[2],
		Cidr:        m[1],
	}
}

func (objMgr *ObjectManager) GetNetwork(netview string, cidr string, ea EA) (*Network, error) {
	var res []Network

	network := NewNetwork(Network{
		NetviewName: netview})

	if cidr != "" {
		network.Cidr = cidr
	}

	if ea != nil && len(ea) > 0 {
		network.eaSearch = EASearch(ea)
	}

	err := objMgr.connector.GetObject(network, "", &res)

	if err != nil || res == nil || len(res) == 0 {
		return nil, err
	}

	return &res[0], nil
}

func (objMgr *ObjectManager) GetNetworkwithref(ref string) (*Network, error) {
	network := NewNetwork(Network{})
	err := objMgr.connector.GetObject(network, ref, &network)
	return network, err
}

func (objMgr *ObjectManager) GetNetworkContainer(netview string, cidr string) (*NetworkContainer, error) {
	var res []NetworkContainer

	nwcontainer := NewNetworkContainer(NetworkContainer{
		NetviewName: netview,
		Cidr:        cidr})

	err := objMgr.connector.GetObject(nwcontainer, "", &res)

	if err != nil || res == nil || len(res) == 0 {
		return nil, err
	}

	return &res[0], nil
}

func GetIPAddressFromRef(ref string) string {
	// fixedaddress/ZG5zLmJpbmRfY25h:12.0.10.1/external
	r := regexp.MustCompile(`fixedaddress/\w+:(\d+\.\d+\.\d+\.\d+)/.+`)
	m := r.FindStringSubmatch(ref)

	if m != nil {
		return m[1]
	}
	return ""
}

func (objMgr *ObjectManager) AllocateIP(netview string, cidr string, ipAddr string, macAddress string, name string, vmID string, vmName string) (*FixedAddress, error) {
	if len(macAddress) == 0 {
		macAddress = MACADDR_ZERO
	}

	ea := objMgr.getBasicVMEA(true, vmID, vmName)
	fixedAddr := NewFixedAddress(FixedAddress{
		NetviewName: netview,
		Cidr:        cidr,
		Mac:         macAddress,
		Name:        name,
		Ea:          ea})

	if ipAddr == "" {
		fixedAddr.IPAddress = fmt.Sprintf("func:nextavailableip:%s,%s", cidr, netview)
	} else {
		fixedAddr.IPAddress = ipAddr
	}

	ref, err := objMgr.connector.CreateObject(fixedAddr)
	fixedAddr.Ref = ref
	fixedAddr.IPAddress = GetIPAddressFromRef(ref)

	return fixedAddr, err
}

func (objMgr *ObjectManager) AllocateNetwork(netview string, cidr string, prefixLen uint, name string) (network *Network, err error) {
	network = nil

	networkReq := NewNetwork(Network{
		NetviewName: netview,
		Cidr:        fmt.Sprintf("func:nextavailablenetwork:%s,%s,%d", cidr, netview, prefixLen),
		Ea:          objMgr.getBasicEA(true)})
	if name != "" {
		networkReq.Ea["Network Name"] = name
	}

	ref, err := objMgr.connector.CreateObject(networkReq)
	if err == nil && len(ref) > 0 {
		network = BuildNetworkFromRef(ref)
	}

	return
}

func (objMgr *ObjectManager) GetFixedAddress(netview string, cidr string, ipAddr string, macAddr string) (*FixedAddress, error) {
	var res []FixedAddress

	fixedAddr := NewFixedAddress(FixedAddress{
		NetviewName: netview,
		Cidr:        cidr,
		IPAddress:   ipAddr})

	if macAddr != "" {
		fixedAddr.Mac = macAddr
	}

	err := objMgr.connector.GetObject(fixedAddr, "", &res)

	if err != nil || res == nil || len(res) == 0 {
		return nil, err
	}

	return &res[0], nil
}

func (objMgr *ObjectManager) GetFixedAddressByRef(ref string) (*FixedAddress, error) {
	fixedAddr := NewFixedAddress(FixedAddress{})
	err := objMgr.connector.GetObject(fixedAddr, ref, &fixedAddr)
	return fixedAddr, err
}

func (objMgr *ObjectManager) DeleteFixedAddress(ref string) (string, error) {
	return objMgr.connector.DeleteObject(ref)
}

// validation  for match_client
func validateMatchClient(value string) bool {
	match_client := [5]string{"MAC_ADDRESS", "CLIENT_ID", "RESERVED", "CIRCUIT_ID", "REMOTE_ID"}

	for _, val := range match_client {
		if val == value {
			return true
		}
	}
	return false
}

func (objMgr *ObjectManager) UpdateFixedAddress(fixedAddrRef string, matchClient string, macAddress string, vmID string, vmName string) (*FixedAddress, error) {
	updateFixedAddr := NewFixedAddress(FixedAddress{Ref: fixedAddrRef})

	if len(macAddress) != 0 {
		updateFixedAddr.Mac = macAddress
	}

	ea := objMgr.getBasicVMEA(true, vmID, vmName)

	updateFixedAddr.Ea = ea

	if matchClient != "" {
		if validateMatchClient(matchClient) {
			updateFixedAddr.MatchClient = matchClient
		} else {
			return nil, fmt.Errorf("wrong value for match_client passed %s \n ", matchClient)
		}
	}

	refResp, err := objMgr.connector.UpdateObject(updateFixedAddr, fixedAddrRef)
	updateFixedAddr.Ref = refResp
	return updateFixedAddr, err
}

func (objMgr *ObjectManager) ReleaseIP(netview string, cidr string, ipAddr string, macAddr string) (string, error) {
	fixAddress, _ := objMgr.GetFixedAddress(netview, cidr, ipAddr, macAddr)
	if fixAddress == nil {
		return "", nil
	}
	return objMgr.connector.DeleteObject(fixAddress.Ref)
}

func (objMgr *ObjectManager) DeleteNetwork(ref string, netview string) (string, error) {
	network := BuildNetworkFromRef(ref)
	if network != nil && network.NetviewName == netview {
		return objMgr.connector.DeleteObject(ref)
	}

	return "", nil
}

func (objMgr *ObjectManager) GetEADefinition(name string) (*EADefinition, error) {
	var res []EADefinition

	eadef := NewEADefinition(EADefinition{Name: name})

	err := objMgr.connector.GetObject(eadef, "", &res)

	if err != nil || res == nil || len(res) == 0 {
		return nil, err
	}

	return &res[0], nil
}

func (objMgr *ObjectManager) CreateEADefinition(eadef EADefinition) (*EADefinition, error) {
	newEadef := NewEADefinition(eadef)

	ref, err := objMgr.connector.CreateObject(newEadef)
	newEadef.Ref = ref

	return newEadef, err
}

func (objMgr *ObjectManager) CreateHostRecord(enabledns bool, recordName string, netview string, dnsview string, cidr string, ipAddr string, macAddress string, vmID string, vmName string) (*HostRecord, error) {

	ea := objMgr.getBasicVMEA(true, vmID, vmName)

	recordHostIpAddr := NewHostRecordIpv4Addr(HostRecordIpv4Addr{Mac: macAddress})

	if ipAddr == "" {
		recordHostIpAddr.Ipv4Addr = fmt.Sprintf("func:nextavailableip:%s,%s", cidr, netview)
	} else {
		recordHostIpAddr.Ipv4Addr = ipAddr
	}
	enableDNS := new(bool)
	*enableDNS = enabledns
	recordHostIpAddrSlice := []HostRecordIpv4Addr{*recordHostIpAddr}
	recordHost := NewHostRecord(HostRecord{
		Name:        recordName,
		EnableDns:   enableDNS,
		NetworkView: netview,
		View:        dnsview,
		Ipv4Addrs:   recordHostIpAddrSlice,
		Ea:          ea})

	ref, err := objMgr.connector.CreateObject(recordHost)
	recordHost.Ref = ref
	err = objMgr.connector.GetObject(recordHost, ref, &recordHost)
	return recordHost, err
}

func (objMgr *ObjectManager) GetHostRecordByRef(ref string) (*HostRecord, error) {
	recordHost := NewHostRecord(HostRecord{})
	err := objMgr.connector.GetObject(recordHost, ref, &recordHost)
	return recordHost, err
}

func (objMgr *ObjectManager) GetHostRecord(recordName string, netview string, cidr string, ipAddr string) (*HostRecord, error) {
	var res []HostRecord

	recordHost := NewHostRecord(HostRecord{})
	if recordName != "" {
		recordHost.Name = recordName
	}

	err := objMgr.connector.GetObject(recordHost, "", &res)

	if err != nil || res == nil || len(res) == 0 {
		return nil, err
	}
	return &res[0], err

}

func (objMgr *ObjectManager) GetIpAddressFromHostRecord(host HostRecord) (string, error) {
	err := objMgr.connector.GetObject(&host, host.Ref, &host)
	return host.Ipv4Addrs[0].Ipv4Addr, err
}

func (objMgr *ObjectManager) UpdateHostRecord(hostRref string, ipAddr string, macAddress string, vmID string, vmName string) (string, error) {

	recordHostIpAddr := NewHostRecordIpv4Addr(HostRecordIpv4Addr{Mac: macAddress, Ipv4Addr: ipAddr})
	recordHostIpAddrSlice := []HostRecordIpv4Addr{*recordHostIpAddr}
	updateHostRecord := NewHostRecord(HostRecord{Ipv4Addrs: recordHostIpAddrSlice})

	ea := objMgr.getBasicVMEA(true, vmID, vmName)

	updateHostRecord.Ea = ea

	ref, err := objMgr.connector.UpdateObject(updateHostRecord, hostRref)
	return ref, err
}

func (objMgr *ObjectManager) DeleteHostRecord(ref string) (string, error) {
	return objMgr.connector.DeleteObject(ref)
}

func (objMgr *ObjectManager) CreateARecord(netview string, dnsview string, recordname string, cidr string, ipAddr string, vmID string, vmName string) (*RecordA, error) {

	ea := objMgr.getBasicVMEA(true, vmID, vmName)

	recordA := NewRecordA(RecordA{
		View: dnsview,
		Name: recordname,
		Ea:   ea})

	if ipAddr == "" {
		recordA.Ipv4Addr = fmt.Sprintf("func:nextavailableip:%s,%s", cidr, netview)
	} else {
		recordA.Ipv4Addr = ipAddr
	}
	ref, err := objMgr.connector.CreateObject(recordA)
	recordA.Ref = ref
	return recordA, err
}

func (objMgr *ObjectManager) GetARecordByRef(ref string) (*RecordA, error) {
	recordA := NewRecordA(RecordA{})
	err := objMgr.connector.GetObject(recordA, ref, &recordA)
	return recordA, err
}

func (objMgr *ObjectManager) DeleteARecord(ref string) (string, error) {
	return objMgr.connector.DeleteObject(ref)
}

func (objMgr *ObjectManager) CreateCNAMERecord(canonical string, recordname string, dnsview string) (*RecordCNAME, error) {

	recordCNAME := NewRecordCNAME(RecordCNAME{
		View:      dnsview,
		Name:      recordname,
		Canonical: canonical})

	ref, err := objMgr.connector.CreateObject(recordCNAME)
	recordCNAME.Ref = ref
	return recordCNAME, err
}

func (objMgr *ObjectManager) GetCNAMERecordByRef(ref string) (*RecordCNAME, error) {
	recordCNAME := NewRecordCNAME(RecordCNAME{})
	err := objMgr.connector.GetObject(recordCNAME, ref, &recordCNAME)
	return recordCNAME, err
}

func (objMgr *ObjectManager) DeleteCNAMERecord(ref string) (string, error) {
	return objMgr.connector.DeleteObject(ref)
}

func (objMgr *ObjectManager) CreatePTRRecord(netview string, dnsview string, recordname string, cidr string, ipAddr string, vmID string, vmName string) (*RecordPTR, error) {

	ea := objMgr.getBasicVMEA(true, vmID, vmName)

	recordPTR := NewRecordPTR(RecordPTR{
		View:     dnsview,
		PtrdName: recordname,
		Ea:       ea})

	if ipAddr == "" {
		recordPTR.Ipv4Addr = fmt.Sprintf("func:nextavailableip:%s,%s", cidr, netview)
	} else {
		recordPTR.Ipv4Addr = ipAddr
	}
	ref, err := objMgr.connector.CreateObject(recordPTR)
	recordPTR.Ref = ref
	return recordPTR, err
}

func (objMgr *ObjectManager) GetPTRRecordByRef(ref string) (*RecordPTR, error) {
	recordPTR := NewRecordPTR(RecordPTR{})
	err := objMgr.connector.GetObject(recordPTR, ref, &recordPTR)
	return recordPTR, err
}

func (objMgr *ObjectManager) DeletePTRRecord(ref string) (string, error) {
	return objMgr.connector.DeleteObject(ref)
}

// CreateMultiObject unmarshals the result into slice of maps
func (objMgr *ObjectManager) CreateMultiObject(req *MultiRequest) ([]map[string]interface{}, error) {

	conn := objMgr.connector.(*Connector)
	queryParams := QueryParams{forceProxy: false}
	res, err := conn.makeRequest(CREATE, req, "", queryParams)

	if err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	err = json.Unmarshal(res, &result)

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetUpgradeStatus returns the grid upgrade information
func (objMgr *ObjectManager) GetUpgradeStatus(statusType string) ([]UpgradeStatus, error) {
	var res []UpgradeStatus

	if statusType == "" {
		// TODO option may vary according to the WAPI version, need to
		// throw relevant  error.
		msg := fmt.Sprintf("Status type can not be nil")
		return res, errors.New(msg)
	}
	upgradestatus := NewUpgradeStatus(UpgradeStatus{Type: statusType})
	err := objMgr.connector.GetObject(upgradestatus, "", &res)

	return res, err
}

// GetAllMembers returns all members information
func (objMgr *ObjectManager) GetAllMembers() ([]Member, error) {
	var res []Member

	memberObj := NewMember(Member{})
	err := objMgr.connector.GetObject(memberObj, "", &res)
	return res, err
}

// GetCapacityReport returns all capacity for members
func (objMgr *ObjectManager) GetCapacityReport(name string) ([]CapacityReport, error) {
	var res []CapacityReport

	capacityObj := CapacityReport{Name: name}
	capacityReport := NewCapcityReport(capacityObj)
	err := objMgr.connector.GetObject(capacityReport, "", &res)
	return res, err
}

// GetLicense returns the license details for member
func (objMgr *ObjectManager) GetLicense() ([]License, error) {
	var res []License

	licenseObj := NewLicense(License{})
	err := objMgr.connector.GetObject(licenseObj, "", &res)
	return res, err
}

// GetLicense returns the license details for grid
func (objMgr *ObjectManager) GetGridLicense() ([]License, error) {
	var res []License

	licenseObj := NewGridLicense(License{})
	err := objMgr.connector.GetObject(licenseObj, "", &res)
	return res, err
}

// GetGridInfo returns the details for grid
func (objMgr *ObjectManager) GetGridInfo() ([]Grid, error) {
	var res []Grid

	gridObj := NewGrid(Grid{})
	err := objMgr.connector.GetObject(gridObj, "", &res)
	return res, err
}
