package docker

import (
	"math"
	"strconv"
	"text/template"

	"github.com/BurntSushi/ty/fun"
	"github.com/containous/traefik/log"
	"github.com/containous/traefik/provider/label"
	"github.com/containous/traefik/types"
)

func (p *Provider) buildConfiguration(containersInspected []dockerData) *types.Configuration {
	var DockerFuncMap = template.FuncMap{
		"getBackend":                  getBackend,
		"getIPAddress":                p.getIPAddress,
		"getPort":                     getPort,
		"getWeight":                   getFuncStringLabel(label.TraefikWeight, label.DefaultWeight),
		"getDomain":                   getFuncStringLabel(label.TraefikDomain, p.Domain),
		"getProtocol":                 getFuncStringLabel(label.TraefikProtocol, label.DefaultProtocol),
		"getPassHostHeader":           getFuncStringLabel(label.TraefikFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getPriority":                 getFuncStringLabel(label.TraefikFrontendPriority, label.DefaultFrontendPriority),
		"getEntryPoints":              getFuncSliceStringLabel(label.TraefikFrontendEntryPoints),
		"getBasicAuth":                getFuncSliceStringLabel(label.TraefikFrontendAuthBasic),
		"getFrontendRule":             p.getFrontendRule,
		"getRedirect":                 getFuncStringLabel(label.TraefikFrontendRedirect, label.DefaultFrontendRedirect),
		"hasCircuitBreakerLabel":      hasFunc(label.TraefikBackendCircuitBreakerExpression),
		"getCircuitBreakerExpression": getFuncStringLabel(label.TraefikBackendCircuitBreakerExpression, label.DefaultCircuitBreakerExpression),
		"hasLoadBalancerLabel":        hasLoadBalancerLabel,
		"getLoadBalancerMethod":       getFuncStringLabel(label.TraefikBackendLoadBalancerMethod, label.DefaultBackendLoadBalancerMethod),
		"hasMaxConnLabels":            hasMaxConnLabels,
		"getMaxConnAmount":            getFuncInt64Label(label.TraefikBackendMaxConnAmount, math.MaxInt64),
		"getMaxConnExtractorFunc":     getFuncStringLabel(label.TraefikBackendMaxConnExtractorFunc, label.DefaultBackendMaxconnExtractorFunc),
		"getSticky":                   getSticky,
		"hasStickinessLabel":          hasFunc(label.TraefikBackendLoadBalancerStickiness),
		"getStickinessCookieName":     getFuncStringLabel(label.TraefikBackendLoadBalancerStickinessCookieName, label.DefaultBackendLoadbalancerStickinessCookieName),
		"isBackendLBSwarm":            isBackendLBSwarm, // FIXME DEAD ?
		"getServiceBackend":           getServiceBackend,
		"getServiceRedirect":          getFuncServiceStringLabel(label.SuffixFrontendRedirect, label.DefaultFrontendRedirect),
		"getWhitelistSourceRange":     getFuncSliceStringLabel(label.TraefikFrontendWhitelistSourceRange),

		"hasRequestHeaders":                 hasFunc(label.TraefikFrontendRequestHeaders),
		"getRequestHeaders":                 getFuncMapLabel(label.TraefikFrontendRequestHeaders),
		"hasResponseHeaders":                hasFunc(label.TraefikFrontendResponseHeaders),
		"getResponseHeaders":                getFuncMapLabel(label.TraefikFrontendResponseHeaders),
		"hasAllowedHostsHeaders":            hasFunc(label.TraefikFrontendAllowedHosts),
		"getAllowedHostsHeaders":            getFuncSliceStringLabel(label.TraefikFrontendAllowedHosts),
		"hasHostsProxyHeaders":              hasFunc(label.TraefikFrontendHostsProxyHeaders),
		"getHostsProxyHeaders":              getFuncSliceStringLabel(label.TraefikFrontendHostsProxyHeaders),
		"hasSSLRedirectHeaders":             hasFunc(label.TraefikFrontendSSLRedirect),
		"getSSLRedirectHeaders":             getFuncBoolLabel(label.TraefikFrontendSSLRedirect, false),
		"hasSSLTemporaryRedirectHeaders":    hasFunc(label.TraefikFrontendSSLTemporaryRedirect),
		"getSSLTemporaryRedirectHeaders":    getFuncBoolLabel(label.TraefikFrontendSSLTemporaryRedirect, false),
		"hasSSLHostHeaders":                 hasFunc(label.TraefikFrontendSSLHost),
		"getSSLHostHeaders":                 getFuncStringLabel(label.TraefikFrontendSSLHost, ""),
		"hasSSLProxyHeaders":                hasFunc(label.TraefikFrontendSSLProxyHeaders),
		"getSSLProxyHeaders":                getFuncMapLabel(label.TraefikFrontendSSLProxyHeaders),
		"hasSTSSecondsHeaders":              hasFunc(label.TraefikFrontendSTSSeconds),
		"getSTSSecondsHeaders":              getFuncInt64Label(label.TraefikFrontendSTSSeconds, 0),
		"hasSTSIncludeSubdomainsHeaders":    hasFunc(label.TraefikFrontendSTSIncludeSubdomains),
		"getSTSIncludeSubdomainsHeaders":    getFuncBoolLabel(label.TraefikFrontendSTSIncludeSubdomains, false),
		"hasSTSPreloadHeaders":              hasFunc(label.TraefikFrontendSTSPreload),
		"getSTSPreloadHeaders":              getFuncBoolLabel(label.TraefikFrontendSTSPreload, false),
		"hasForceSTSHeaderHeaders":          hasFunc(label.TraefikFrontendForceSTSHeader),
		"getForceSTSHeaderHeaders":          getFuncBoolLabel(label.TraefikFrontendForceSTSHeader, false),
		"hasFrameDenyHeaders":               hasFunc(label.TraefikFrontendFrameDeny),
		"getFrameDenyHeaders":               getFuncBoolLabel(label.TraefikFrontendFrameDeny, false),
		"hasCustomFrameOptionsValueHeaders": hasFunc(label.TraefikFrontendCustomFrameOptionsValue),
		"getCustomFrameOptionsValueHeaders": getFuncStringLabel(label.TraefikFrontendCustomFrameOptionsValue, ""),
		"hasContentTypeNosniffHeaders":      hasFunc(label.TraefikFrontendContentTypeNosniff),
		"getContentTypeNosniffHeaders":      getFuncBoolLabel(label.TraefikFrontendContentTypeNosniff, false),
		"hasBrowserXSSFilterHeaders":        hasFunc(label.TraefikFrontendBrowserXSSFilter),
		"getBrowserXSSFilterHeaders":        getFuncBoolLabel(label.TraefikFrontendBrowserXSSFilter, false),
		"hasContentSecurityPolicyHeaders":   hasFunc(label.TraefikFrontendContentSecurityPolicy),
		"getContentSecurityPolicyHeaders":   getFuncStringLabel(label.TraefikFrontendContentSecurityPolicy, ""),
		"hasPublicKeyHeaders":               hasFunc(label.TraefikFrontendPublicKey),
		"getPublicKeyHeaders":               getFuncStringLabel(label.TraefikFrontendPublicKey, ""),
		"hasReferrerPolicyHeaders":          hasFunc(label.TraefikFrontendReferrerPolicy),
		"getReferrerPolicyHeaders":          getFuncStringLabel(label.TraefikFrontendReferrerPolicy, ""),
		"hasIsDevelopmentHeaders":           hasFunc(label.TraefikFrontendIsDevelopment),
		"getIsDevelopmentHeaders":           getFuncBoolLabel(label.TraefikFrontendIsDevelopment, false),

		"hasServices":               hasServices,
		"getServiceNames":           getServiceNames,
		"getServicePort":            getServicePort,
		"hasServiceRequestHeaders":  hasFuncServiceLabel(label.SuffixFrontendRequestHeaders),
		"getServiceRequestHeaders":  getFuncServiceMapLabel(label.SuffixFrontendRequestHeaders),
		"hasServiceResponseHeaders": hasFuncServiceLabel(label.SuffixFrontendResponseHeaders),
		"getServiceResponseHeaders": getFuncServiceMapLabel(label.SuffixFrontendResponseHeaders),
		"getServiceWeight":          getFuncServiceStringLabel(label.SuffixWeight, label.DefaultWeight),
		"getServiceProtocol":        getFuncServiceStringLabel(label.SuffixProtocol, label.DefaultProtocol),
		"getServiceEntryPoints":     getFuncServiceSliceStringLabel(label.SuffixFrontendEntryPoints),
		"getServiceBasicAuth":       getFuncServiceSliceStringLabel(label.SuffixFrontendAuthBasic),
		"getServiceFrontendRule":    p.getServiceFrontendRule,
		"getServicePassHostHeader":  getFuncServiceStringLabel(label.SuffixFrontendPassHostHeader, label.DefaultPassHostHeader),
		"getServicePriority":        getFuncServiceStringLabel(label.SuffixFrontendPriority, label.DefaultFrontendPriority),
	}
	// filter containers
	filteredContainers := fun.Filter(func(container dockerData) bool {
		return p.containerFilter(container)
	}, containersInspected).([]dockerData)

	frontends := map[string][]dockerData{}
	backends := map[string]dockerData{}
	servers := map[string][]dockerData{}
	serviceNames := make(map[string]struct{})
	for idx, container := range filteredContainers {
		if _, exists := serviceNames[container.ServiceName]; !exists {
			frontendName := p.getFrontendName(container, idx)
			frontends[frontendName] = append(frontends[frontendName], container)
			if len(container.ServiceName) > 0 {
				serviceNames[container.ServiceName] = struct{}{}
			}
		}
		backendName := getBackend(container)
		backends[backendName] = container
		servers[backendName] = append(servers[backendName], container)
	}

	templateObjects := struct {
		Containers []dockerData
		Frontends  map[string][]dockerData
		Backends   map[string]dockerData
		Servers    map[string][]dockerData
		Domain     string
	}{
		filteredContainers,
		frontends,
		backends,
		servers,
		p.Domain,
	}

	configuration, err := p.GetConfiguration("templates/docker.tmpl", DockerFuncMap, templateObjects)
	if err != nil {
		log.Error(err)
	}

	return configuration
}

func (p Provider) containerFilter(container dockerData) bool {
	if !label.IsEnabled(container.Labels, p.ExposedByDefault) {
		log.Debugf("Filtering disabled container %s", container.Name)
		return false
	}

	var err error
	portLabel := "traefik.port label"
	if hasServices(container) {
		portLabel = "traefik.<serviceName>.port or " + portLabel + "s"
		err = checkServiceLabelPort(container)
	} else {
		_, err = strconv.Atoi(container.Labels[label.TraefikPort])
	}
	if len(container.NetworkSettings.Ports) == 0 && err != nil {
		log.Debugf("Filtering container without port and no %s %s : %s", portLabel, container.Name, err.Error())
		return false
	}

	constraintTags := label.SplitAndTrimString(container.Labels[label.TraefikTags], ",")
	if ok, failingConstraint := p.MatchConstraints(constraintTags); !ok {
		if failingConstraint != nil {
			log.Debugf("Container %v pruned by '%v' constraint", container.Name, failingConstraint.String())
		}
		return false
	}

	if container.Health != "" && container.Health != "healthy" {
		log.Debugf("Filtering unhealthy or starting container %s", container.Name)
		return false
	}

	if len(p.getFrontendRule(container)) == 0 {
		log.Debugf("Filtering container with empty frontend rule %s", container.Name)
		return false
	}

	return true
}
