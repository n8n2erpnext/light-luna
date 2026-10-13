# 13 — Research Sources and Verification Notes

Reviewed 2026-10-13. Upstream endpoints, editions, release behavior and contract terms change over time. **Do not treat a design assumption as verified deployment behavior.** Revalidate against the exact installed versions before implementing.

## NetBird official documentation

- Architecture: https://docs.netbird.io/about-netbird/how-netbird-works
- API intro: https://docs.netbird.io/api/introduction
- Authentication/service users: https://docs.netbird.io/api/guides/authentication and https://docs.netbird.io/manage/public-api
- Peers API: https://docs.netbird.io/api/resources/peers
- Groups API: https://docs.netbird.io/api/resources/groups
- Policies API: https://docs.netbird.io/api/resources/policies
- Networks/Resources/Routers API: https://docs.netbird.io/api/resources/networks
- Events API: https://docs.netbird.io/api/resources/events
- Host firewall conflicts, `wt0`: https://docs.netbird.io/about-netbird/ports-and-firewalls
- Client environment and table name: https://docs.netbird.io/client/environment-variables
- Client firewall consequences: https://docs.netbird.io/help/troubleshooting-client
- Networks masquerade: https://docs.netbird.io/manage/networks/masquerade
- Deprecated Routes context: https://docs.netbird.io/manage/network-routes
- Reverse Proxy: https://docs.netbird.io/manage/reverse-proxy
- Proxy access/CrowdSec: https://docs.netbird.io/selfhosted/maintenance/crowdsec
- Self-hosted architecture and proxy variants: https://docs.netbird.io/selfhosted/selfhosted-guide and https://docs.netbird.io/selfhosted/external-reverse-proxy

## Firewall/WAF source references

- nftables timed set elements: https://wiki.nftables.org/wiki-nftables/index.php/Element_timeouts
- nftables project/documentation: https://wiki.nftables.org/
- Coraza WAF Go integration: https://www.coraza.io/docs/tutorials/quick-start/
- OWASP Core Rule Set: https://coreruleset.org/
- Suricata documentation: https://docs.suricata.io/
- CISA KEV: https://www.cisa.gov/known-exploited-vulnerabilities-catalog
- NVD APIs: https://nvd.nist.gov/developers/vulnerabilities
- FIRST EPSS: https://www.first.org/epss/
- SPDX: https://spdx.dev/ and CycloneDX: https://cyclonedx.org/

## Observations from prior Sentinel research (not independently revalidated here)

Luna/Selene existing lab can collect and aggregate site-local observations and issue shadow decisions without real bans. Known gaps include independent original-client-IP proof, complete deployed app SBOM/CVE applicability and independently labeled detection evaluation. These are explicit initial unknowns; they do not establish the Light Luna firewall implementation or product-ready AI detection.
