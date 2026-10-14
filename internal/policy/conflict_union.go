package policy

// This file resolves finite CIDR-set coverage for a static rule audit. It is
// pure Go arithmetic: no sockets, kernel inspection, nft or policy execution.
import (
	"net/netip"
	"sort"
)

type prefixInterval struct{ first, last netip.Addr }

// prefixLast returns the inclusive last address of a canonical prefix.
// Prefix parsing/canonicality is validated by Bundle.Validate beforehand.
func prefixLast(p netip.Prefix) netip.Addr {
	if p.Addr().Is4() {
		bits := p.Addr().As4()
		for i := p.Bits(); i < 32; i++ {
			bits[i/8] |= 1 << uint(7-i%8)
		}
		return netip.AddrFrom4(bits)
	}
	bits := p.Addr().As16()
	for i := p.Bits(); i < 128; i++ {
		bits[i/8] |= 1 << uint(7-i%8)
	}
	return netip.AddrFrom16(bits)
}

// unionContainsPrefix proves every address in target belongs to at least one
// member of fragments. It returns false on any uncovered gap. Every fragment
// is a canonical subnet of target of the same IP family. No false-positive
// coverage claims are allowed.
func unionContainsPrefix(target netip.Prefix, fragments []netip.Prefix) bool {
	if len(fragments) == 0 {
		return false
	}
	intervals := make([]prefixInterval, 0, len(fragments))
	for _, part := range fragments {
		if !prefixContains(target, part) {
			continue
		}
		intervals = append(intervals, prefixInterval{first: part.Addr(), last: prefixLast(part)})
	}
	if len(intervals) == 0 {
		return false
	}
	sort.Slice(intervals, func(i, j int) bool {
		if c := intervals[i].first.Compare(intervals[j].first); c != 0 {
			return c < 0
		}
		return intervals[i].last.Compare(intervals[j].last) > 0
	})
	cursor := target.Addr()
	end := prefixLast(target)
	for _, part := range intervals {
		if part.first.Compare(cursor) > 0 {
			return false
		} // uncovered gap
		if part.last.Compare(end) >= 0 {
			return true
		}
		if part.last.Compare(cursor) >= 0 {
			next := part.last.Next()
			if !next.IsValid() {
				return false
			}
			cursor = next
		}
	}
	return false
}
