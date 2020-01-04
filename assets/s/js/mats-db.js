var db;
if (!db) { db = {}; }
db.matKinds = {
    raw: {ord: 1, sym: "⎔"},
    man: {ord: 2, sym: "⚙"},
    enc: {ord: 3, sym: "⛁"}
};
db.mats = {
    rSb: {kind: "raw", grade: 4},
    rAs: {kind: "raw", grade: 2},
    rB:  {kind: "raw", grade: 1},
    rCd: {kind: "raw", grade: 3},
    rC:  {kind: "raw", grade: 1},
    rCr: {kind: "raw", grade: 2},
    rGe: {kind: "raw", grade: 2},
    rFe: {kind: "raw", grade: 1},
    rPb: {kind: "raw", grade: 1},
    rMn: {kind: "raw", grade: 2},
    rHg: {kind: "raw", grade: 3},
    rMo: {kind: "raw", grade: 3},
    rNi: {kind: "raw", grade: 1},
    rNb: {kind: "raw", grade: 3},
    rP:  {kind: "raw", grade: 1},
    rPo: {kind: "raw", grade: 4},
    rRe: {kind: "raw", grade: 1},
    rRu: {kind: "raw", grade: 4},
    rSe: {kind: "raw", grade: 4},
    rS:  {kind: "raw", grade: 1},
    rTc: {kind: "raw", grade: 4},
    rTe: {kind: "raw", grade: 4},
    rSn: {kind: "raw", grade: 3},
    rW:  {kind: "raw", grade: 3},
    rV:  {kind: "raw", grade: 2},
    rY:  {kind: "raw", grade: 4},
    rZn: {kind: "raw", grade: 3},
    rZr: {kind: "raw", grade: 4},

    mBscCndc: {kind: "man", grade: 1},
    mBioCndc: {kind: "man", grade: 5},
    mChmDist: {kind: "man", grade: 3},
    mChmMnp: {kind: "man", grade: 4},
    mChmPrc: {kind: "man", grade: 2},
    mChmStrU: {kind: "man", grade: 1},
    mComCmps: {kind: "man", grade: 1},
    mCmpShld: {kind: "man", grade: 4},
    mCndCerm: {kind: "man", grade: 3},
    mCndComp: {kind: "man", grade: 2},
    mCndPoly: {kind: "man", grade: 4},
    mCnfComp: {kind: "man", grade: 4},
    mCDCmps: {kind: "man", grade: 5},
    mCrysSh: {kind: "man", grade: 1},
    mElcAry: {kind: "man", grade: 3},
    mExqFCrs: {kind: "man", grade: 5},
    mFilComp: {kind: "man", grade: 2},
    mFlwFCrs: {kind: "man", grade: 2},
    mFCrs: {kind: "man", grade: 3},
    mGlvAly: {kind: "man", grade: 2},
    mGrdRes: {kind: "man", grade: 1},
    mHtCndWr: {kind: "man", grade: 1},
    mHtDspP: {kind: "man", grade: 2},
    mHtXchg: {kind: "man", grade: 3},
    mHtRsCrm: {kind: "man", grade: 2},
    mHtVns: {kind: "man", grade: 4},
    mHiDnsComp: {kind: "man", grade: 3},
    mHybCpct: {kind: "man", grade: 2},
    mImpShld: {kind: "man", grade: 5},
    mIprComp: {kind: "man", grade: 5},
    mMchComp: {kind: "man", grade: 3},
    mMchEqp: {kind: "man", grade: 2},
    mMchScr: {kind: "man", grade: 1},
    mMilGrAly: {kind: "man", grade: 5},
    mMilSCap: {kind: "man", grade: 5},
    mPhIso: {kind: "man", grade: 5},
    mPhsAly: {kind: "man", grade: 3},
    mPlyCap: {kind: "man", grade: 4},

    dAbrSPA: {kind: "enc", grade: 4},
    dAbnCED: {kind: "enc", grade: 5},
    dAdpEC: {kind: "enc", grade: 5},
    dAnmBSD: {kind: "enc", grade: 1},
    dAnmFSDT: {kind: "enc", grade: 2},
    dAtpDWE: {kind: "enc", grade: 1},
    dAtpEncAr: {kind: "enc", grade: 4},
    dClfScDta: {kind: "enc", grade: 3},
    
    dIrrEmDta: {kind: "enc", grade: 2}
};
db.rcps = {
    sSrvFuel1: {
		kind: "synth",
		mats: {rS: 2, rP: 1}
    },
    sSrvFuel2: {
		kind: "synth",
		mats: {rP: 1, rMn: 1, rSe: 1, rMo: 1}
    },
    sSrvFuel3: {
		kind: "synth",
		mats: {rP: 2, rSe: 2, rMo: 1, rTc: 1}
    },
	
    tGFsd: {
		kind: "gtech",
		mats: {mChmDist: 1, dIrrEmDta: 3}
    },

    eFsdRng1: {kind: "eng",
	       mats: {dAtpDWE: 1}},
    eFsdRng2: {kind: "eng",
	       mats: {dAtpDWE: 1, mChmPrc:1}},
    eFsdRng3: {kind: "eng",
	       mats: {rP: 1, mChmPrc:1, dSrgWkSol: 1}},
    eFsdRng4: {kind: "eng",
	       mats: {rMn: 1, mChmDist: 1, dEccHypTr: 1}},
    eFsdRng5: {kind: "eng",
	       mats: {rAs: 1, mChmMnp: 1, dDmWkXpt: 1}},
};
