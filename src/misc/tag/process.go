package tag

import (
	"github.com/skycoin/bbs/src/misc/boo"
	"github.com/skycoin/bbs/src/misc/keys"
	"github.com/skycoin/skycoin/src/cipher"
	"regexp"
	"strconv"
	"strings"
)

const (
	processKey      = "bbs"
	valUserPKStr    = "upkStr"
	valUserPK       = "upk"
	valUserSK       = "usk"
	valUserRefStr   = "uRefStr"
	valUserRef      = "uRef"
	valBoardPKStr   = "bpkStr"
	valBoardPK      = "bpk"
	valBoardSK      = "bsk"
	valThreadRefStr = "tRefStr"
	valThreadRef    = "tRef"
	valPostRefStr   = "pRefStr"
	valPostRef      = "pRef"
	valBoardSeed    = "bSeed"
	valUserSeed     = "uSeed"
	valHeading      = "heading"
	valBody         = "body"
	valAlias        = "alias"
	valPassword     = "password"
	valAddress      = "address"
	valSubPKsStr    = "subPKsStr"
	valSubPKs       = "subPKs"
	valConsStr      = "consStr"
	valCons         = "cons"
	valModeStr      = "modeStr"
	valMode         = "mode"
	valTagStr       = "tagStr"
	valTag          = "tag"
)

func MultiProcess(objects ...interface{}) error {
	for _, obj := range objects {
		if e := Process(obj); e != nil {
			return e
		}
	}
	return nil
}

// Process processes the fields tagged.
func Process(obj interface{}) error {
	rVal, rTyp := getReflectPair(obj)
	if e := process(makeTagMap(processKey, rVal, rTyp)); e != nil {
		return e
	}
	return nil
}

func process(tm tMap) error {
	// User public key.
	if upkStr, has := tm[valUserPKStr]; has {
		upk, e := keys.GetPubKey(upkStr.String())
		if e != nil {
			return wrapErr(e, "user public key")
		}
		tm.set(valUserPK, upk)
	}
	// User reference.
	if uRefStr, has := tm[valUserRefStr]; has {
		uRef, e := keys.GetPubKey(uRefStr.String())
		if e != nil {
			return wrapErr(e, "user public key reference")
		}
		tm.set(valUserRef, uRef)
	}
	// Board public key.
	if bpkStr, has := tm[valBoardPKStr]; has {
		bpk, e := keys.GetPubKey(bpkStr.String())
		if e != nil {
			return wrapErr(e, "board public key")
		}
		tm.set(valBoardPK, bpk)
	}
	// Thread reference.
	if tRefStr, has := tm[valThreadRefStr]; has {
		tRef, e := keys.GetHash(tRefStr.String())
		if e != nil {
			return wrapErr(e, "thread reference")
		}
		tm.set(valThreadRef, tRef)
	}
	// Post reference.
	if pRefStr, has := tm[valPostRefStr]; has {
		if pRefStr.String() != "" {
			pRef, e := keys.GetHash(pRefStr.String())
			if e != nil {
				return wrapErr(e, "post reference")
			}
			tm.set(valPostRef, pRef)
		}
	}
	// Submission public keys.
	if subPKsStr, has := tm[valSubPKsStr]; has {
		subPKStrs, e := splitStr(subPKsStr.String(), func(v string) bool {
			return true
		})
		if e != nil {
			return wrapErr(e, "submission addresses")
		}
		subPKs := make([]cipher.PubKey, len(subPKStrs))
		for i, pkStr := range subPKStrs {
			if subPKs[i], e = keys.GetPubKey(pkStr); e != nil {
				return wrapErr(e, "submission public keys")
			}
		}
		tm.set(valSubPKs, subPKs)
	}
	// Connections.
	if consStr, has := tm[valConsStr]; has {
		cons, e := splitStr(consStr.String(), func(v string) bool {
			return true
		})
		if e != nil {
			return wrapErr(e, "connections")
		}
		tm.set(valCons, cons)
	}
	// Vote mode.
	if modeStr, has := tm[valModeStr]; has {
		mode, e := strconv.Atoi(modeStr.String())
		if e != nil {
			return boo.WrapType(e, boo.InvalidInput,
				"invalid vote mode input")
		}
		switch int8(mode) {
		case -1, 0, +1:
		default:
			return boo.New(boo.InvalidInput,
				"invalid vote mode input")
		}
		tm.set(valMode, int8(mode))
	}
	// Vote tag.
	if tagStr, has := tm[valTagStr]; has {
		tag := tagStr.String()
		//switch tag {
		//case "", "spam":
		//default:
		//	return boo.New(boo.InvalidInput,
		//		"invalid vote tag input")
		//}
		tm.set(valTag, []byte(tag))
	}
	// User seed.
	if uSeed, has := tm[valUserSeed]; has {
		if e := CheckSeed(uSeed.String()); e != nil {
			return e
		}
		pk, sk := cipher.GenerateDeterministicKeyPair(
			[]byte(uSeed.String()))
		tm.set(valUserPK, pk)
		tm.set(valUserSK, sk)
	}
	// Board seed.
	if bSeed, has := tm[valBoardSeed]; has {
		if e := CheckSeed(bSeed.String()); e != nil {
			return e
		}
		pk, sk := cipher.GenerateDeterministicKeyPair(
			[]byte(bSeed.String()))
		tm.set(valBoardPK, pk)
		tm.set(valBoardSK, sk)
	}
	// Heading text.
	if heading, has := tm[valHeading]; has {
		if e := CheckHeading(heading.String()); e != nil {
			return e
		}
	}
	// Body text.
	if body, has := tm[valBody]; has {
		if e := CheckBody(body.String()); e != nil {
			return e
		}
	}
	// Alias text.
	if alias, has := tm[valAlias]; has {
		if e := CheckAlias(alias.String()); e != nil {
			return e
		}
	}
	// Password text.
	if password, has := tm[valPassword]; has {
		if e := CheckPassword(password.String()); e != nil {
			return e
		}
	}
	// Address text.
	if address, has := tm[valAddress]; has {
		if e := CheckAddress(address.String()); e != nil {
			return e
		}
	}
	return nil
}

/*
	<<< CHECKING FUNCTIONS >>>
*/

// CheckSeed ensures validity of seed. TODO
func CheckSeed(seed string) error {
	return nil
}

// CheckHeading ensures validity of board/thread/post name. TODO
func CheckHeading(heading string) error {
	return nil
}

// CheckBody ensures validity of board/thread/post description. TODO
func CheckBody(body string) error {
	return nil
}

// CheckAlias ensures validity of user alias. TODO
func CheckAlias(alias string) error {
	if len(alias) < 5 {
		return boo.Newf(boo.InvalidInput,
			"alias '%s' is too short", string(alias))
	}
	if len(alias) > 40 {
		return boo.Newf(boo.InvalidInput,
			"alias '%s' is too long", string(alias))
	}
	re := regexp.MustCompile(`^[a-zA-Z0-9]+([._]?[a-zA-Z0-9]+)*$`)
	if !re.Match([]byte(alias)) {
		return boo.Newf(boo.InvalidInput,
			"alias '%s' is invalid", string(alias))
	}

	return nil
}

// CheckPassword ensures validity of password. TODO
func CheckPassword(password string) error {
	return nil
}

// CheckAddress ensures validity of address. TODO
func CheckAddress(address string) error {
	pts := strings.Split(address, ":")
	port, err := strconv.ParseUint(pts[len(pts)-1], 10, 16)
	if err != nil {
		return boo.Newf(boo.InvalidInput,
			"address '%s' is invalid", string(address))
	}
	if port < 0 || port > 65535 {
		return boo.Newf(boo.InvalidInput,
			"address:%s port is invalid ", string(address))
	}
	return nil
}

// CheckMode check's the vote's mode.
func CheckMode(mode int8) error {
	switch mode {
	case -1, 0, +1:
		return nil
	default:
		return boo.Newf(boo.InvalidInput,
			"invalid vote mode of %d provided", mode)
	}
}

// CheckTag check's the vote's tag.
func CheckTag(tag []byte) error {
	switch string(tag) {
	case "", "spam":
		return nil
	default:
		return boo.Newf(boo.InvalidInput,
			"invalid vote tag of %s provided", string(tag))
	}
}
