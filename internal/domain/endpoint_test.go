package domain_test

import(
	"testing"
	"github.com/azargarov/rsvpck/internal/domain"
)

//Status.String/IsSuccess/IsTerminal.
//ProxyConfig.Set/Enabled/MustUseProxy/IsValid (good/bad URLs).
//Endpoint.* validation + IsDirectType/IsProxyType/Key.
//Errorf + IsErrorCode.

func Test_NewEndpoint(t *testing.T) {
	res, err := domain.NewDNSEndpoint("example.com", domain.EndpointTypePublic, "Test DNS endpoint") 
	if err != nil || !res.IsDNS(){
		t.Fatalf("want DNS endpoint got %v, %v", res, err)
	}
	
	res, err = domain.NewICMPEndpoint("example.com", domain.EndpointTypePublic, "Test ICMP endpoint") 
	if err != nil || !res.IsICMP(){
		t.Fatalf("want ICMP endpoint got %v, %v", res, err)
	}
	
	res, err = domain.NewHTTPEndpoint("http://example.com", domain.EndpointTypePublic, "Test HTTP endpoint") 
	if err != nil || !res.IsHTTP(){
		t.Fatalf("want HTTP endpoint got %v, %v", res, err)
	}
	
	res, err = domain.NewTCPEndpoint("  example.com:443", domain.EndpointTypePublic, "Test TCP endpoint") 
	if err != nil || !res.IsTCP(){
		t.Fatalf("want TCP endpoint got %v, %v", res, err)
	}

}