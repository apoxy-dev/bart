// Copyright (c) 2026 Karl Gaissmaier
// SPDX-License-Identifier: MIT

package bart

import (
	"fmt"
	"iter"
	"math/rand/v2"
	"net/netip"
	"testing"
)

// #####################################################
//  this file contains helpers for other test functions
// #####################################################

// workLoadN to adjust loops for tests with -short
func workLoadN() int {
	if testing.Short() {
		return 100
	}
	return 1_000
}

// abbreviation
var mpa = netip.MustParseAddr

// abbreviation and panic on non masked input
var mpp = func(s string) netip.Prefix {
	pfx := netip.MustParsePrefix(s)
	if pfx == pfx.Masked() {
		return pfx
	}
	panic(fmt.Sprintf("%s is not canonicalized as %s", s, pfx.Masked()))
}

// tests for deep copies with Clone method
type MyInt int

// implement the Clone method
func (i *MyInt) Clone() *MyInt {
	if i == nil {
		return i
	}
	a := *i
	return &a
}

func countDumpListNodes[V any](nodes []DumpListNode[V]) int {
	count := len(nodes)
	for _, node := range nodes {
		count += countDumpListNodes(node.Subnets)
	}
	return count
}

func verifyAllIPv4Nodes[V any](t *testing.T, nodes []DumpListNode[V]) {
	for i, node := range nodes {
		if !node.CIDR.Addr().Is4() {
			t.Errorf("Node %d is not IPv4 prefix: %v", i, node.CIDR)
		}
		// Recursively check subnets
		verifyAllIPv4Nodes(t, node.Subnets)
	}
}

func verifyAllIPv6Nodes[V any](t *testing.T, nodes []DumpListNode[V]) {
	for i, node := range nodes {
		if !node.CIDR.Addr().Is6() {
			t.Errorf("Node %d is not IPv6 prefix: %v", i, node.CIDR)
		}
		// Recursively check subnets
		verifyAllIPv6Nodes(t, node.Subnets)
	}
}

func mustPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("%s must panic", name)
		}
	}()
	fn()
}

func noPanic(t *testing.T, name string, fn func()) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("%s panicked: %v", name, r)
		}
	}()
	fn()
}

func mustPanicRangeOverFunc[V any](t *testing.T, name string, fn any) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Fatalf("%s must panic", name)
		}
	}()

	switch iterFunc := fn.(type) {
	case func() iter.Seq[netip.Prefix]:
		for range iterFunc() {
		}
	case func() iter.Seq2[netip.Prefix, V]:
		for range iterFunc() {
		}
	case func(netip.Prefix) iter.Seq[netip.Prefix]:
		pfx := mpp("1.2.3.4/32")
		for range iterFunc(pfx) {
		}
	case func(netip.Prefix) iter.Seq2[netip.Prefix, V]:
		pfx := mpp("1.2.3.4/32")
		for range iterFunc(pfx) {
		}
	default:
		t.Fatalf("%s unknown iter function: %T", name, iterFunc)
	}
}

// randomHostRoutes returns n distinct host routes, half /32 in 10.0.0.0/16 and
// half /128 in 2001:db8::/112. The trie gets deep nodes and many fringes.
func randomHostRoutes(prng *rand.Rand, n int) []netip.Prefix {
	set := make(map[netip.Prefix]bool, n)
	pfxs := make([]netip.Prefix, 0, n)

	for len(pfxs) < n {
		a, b := byte(prng.UintN(256)), byte(prng.UintN(256))

		ip := netip.AddrFrom4([4]byte{10, 0, a, b})
		if len(pfxs)%2 == 1 {
			ip = netip.AddrFrom16([16]byte{0x20, 0x01, 0x0d, 0xb8, 14: a, 15: b})
		}

		pfx := netip.PrefixFrom(ip, ip.BitLen())
		if !set[pfx] {
			set[pfx] = true
			pfxs = append(pfxs, pfx)
		}
	}

	return pfxs
}
