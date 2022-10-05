// NXP i.MX6UL configuration and support
// https://github.com/usbarmory/tamago
//
// Copyright (c) WithSecure Corporation
// https://foundry.withsecure.com
//
// Use of this source code is governed by the license
// that can be found in the LICENSE file.

// Package imx6ul provides support to Go bare metal unikernels, written using
// the TamaGo framework, on the NXP i.MX6UL family of System-on-Chip (SoC)
// application processors.
//
// The package implements initialization and drivers for NXP
// i.MX6UL/i.MX6ULL/i.MX6ULZ SoCs, adopting the following reference
// specifications:
//   * IMX6ULCEC  - i.MX6UL  Data Sheet                               - Rev 2.2 2015/05
//   * IMX6ULLCEC - i.MX6ULL Data Sheet                               - Rev 1.2 2017/11
//   * IMX6ULZCEC - i.MX6ULZ Data Sheet                               - Rev 0   2018/09
//   * IMX6ULRM   - i.MX 6UL  Applications Processor Reference Manual - Rev 1   2016/04
//   * IMX6ULLRM  - i.MX 6ULL Applications Processor Reference Manual - Rev 1   2017/11
//   * IMX6ULZRM  - i.MX 6ULZ Applications Processor Reference Manual - Rev 0   2018/10
//
// This package is only meant to be used with `GOOS=tamago GOARCH=arm` as
// supported by the TamaGo framework for bare metal Go on ARM SoCs, see
// https://github.com/usbarmory/tamago.
package imx6ul

import (
	"encoding/binary"

	"github.com/usbarmory/tamago/arm"
	"github.com/usbarmory/tamago/arm/tzc380"
	"github.com/usbarmory/tamago/internal/reg"
	"github.com/usbarmory/tamago/soc/nxp/csu"
	"github.com/usbarmory/tamago/soc/nxp/dcp"
	"github.com/usbarmory/tamago/soc/nxp/gpio"
	"github.com/usbarmory/tamago/soc/nxp/i2c"
	"github.com/usbarmory/tamago/soc/nxp/ocotp"
	"github.com/usbarmory/tamago/soc/nxp/rngb"
	"github.com/usbarmory/tamago/soc/nxp/snvs"
	"github.com/usbarmory/tamago/soc/nxp/uart"
	"github.com/usbarmory/tamago/soc/nxp/usb"
	"github.com/usbarmory/tamago/soc/nxp/usdhc"
)

// Peripheral registers
const (
	// Central Security Unit
	CSU_BASE = 0x021c0000

	// Data Co-Processor (ULL/ULZ only)
	DCP_BASE = 0x02280000

	// General Interrupt Controller
	GIC_BASE = 0x00a00000

	// General Purpose I/O
	GPIO1_BASE = 0x0209c000
	GPIO2_BASE = 0x020a0000
	GPIO3_BASE = 0x020a4000
	GPIO4_BASE = 0x020a8000
	GPIO5_BASE = 0x020ac000

	// I2C
	I2C1_BASE = 0x021a0000
	I2C2_BASE = 0x021a4000

	// On-Chip OTP Controller
	OCOTP_BASE      = 0x021bc000
	OCOTP_BANK_BASE = 0x021bc400

	// On-Chip Random-Access Memory
	OCRAM_START = 0x00900000
	OCRAM_SIZE  = 0x20000

	// True Random Number Generator (ULL/ULZ only)
	RNGB_BASE = 0x02284000

	// Secure Non-Volatile Storage
	SNVS_BASE = 0x020cc000

	// TrustZone Address Space Controller
	TZASC_BASE            = 0x021d0000
	TZASC_BYPASS          = 0x020e4024
	IOMUXC_GPR_GPR1       = 0x020e4004
	GPR1_TZASC1_BOOT_LOCK = 23

	// Serial ports
	UART1_BASE = 0x02020000
	UART2_BASE = 0x021e8000
	UART3_BASE = 0x021ec000
	UART4_BASE = 0x021f0000

	// USB 2.0
	USB_ANALOG1_BASE   = 0x020c81a0
	USB_ANALOG2_BASE   = 0x020c8200
	USB_ANALOG_DIGPROG = 0x020c8260
	USBPHY1_BASE       = 0x020c9000
	USBPHY2_BASE       = 0x020ca000
	USB1_BASE          = 0x02184000
	USB2_BASE          = 0x02184200

	// SD/MMC
	USDHC1_BASE = 0x02190000
	USDHC2_BASE = 0x02194000
)

// Peripheral instances
var (
	// ARM core
	ARM = &arm.CPU{}

	// Central Security Unit
	CSU = &csu.CSU{
		Base: CSU_BASE,
		CCGR: CCM_CCGR1,
		CG:   CCGRx_CG14,
	}

	// Data Co-Processor (ULL/ULZ only)
	DCP = &dcp.DCP{
		Base: DCP_BASE,
		// DeriveKeyMemory is assigned in init.go
	}

	// GPIO controller 1
	GPIO1 = &gpio.GPIO{
		Index: 1,
		Base:  GPIO1_BASE,
	}

	// GPIO controller 2
	GPIO2 = &gpio.GPIO{
		Index: 2,
		Base:  GPIO2_BASE,
	}

	// GPIO controller 3
	GPIO3 = &gpio.GPIO{
		Index: 3,
		Base:  GPIO3_BASE,
	}

	// GPIO controller 4
	GPIO4 = &gpio.GPIO{
		Index: 4,
		Base:  GPIO4_BASE,
	}

	// GPIO controller 5
	GPIO5 = &gpio.GPIO{
		Index: 5,
		Base:  GPIO5_BASE,
	}

	// I2C controller 1
	I2C1 = &i2c.I2C{
		Index: 1,
		Base:  I2C1_BASE,
		CCGR:  CCM_CCGR2,
		CG:    CCGRx_CG3,
	}

	// I2C controller 2
	I2C2 = &i2c.I2C{
		Index: 2,
		Base:  I2C2_BASE,
		CCGR:  CCM_CCGR2,
		CG:    CCGRx_CG5,
	}

	// On-Chip OTP Controller
	OCOTP = &ocotp.OCOTP{
		Base:     OCOTP_BASE,
		BankBase: OCOTP_BANK_BASE,
		CCGR:     CCM_CCGR2,
		CG:       CCGRx_CG6,
	}

	// True Random Number Generator (ULL/ULZ only)
	RNGB = &rngb.RNGB{
		Base: RNGB_BASE,
	}

	// Secure Non-Volatile Storage
	SNVS = &snvs.SNVS{
		Base: SNVS_BASE,
	}

	// TrustZone Address Space Controller
	TZASC = &tzc380.TZASC{
		Base:              TZASC_BASE,
		Bypass:            TZASC_BYPASS,
		SecureBootLockReg: IOMUXC_GPR_GPR1,
		SecureBootLockPos: GPR1_TZASC1_BOOT_LOCK,
	}

	// Serial port 1
	UART1 = &uart.UART{
		Index: 1,
		Base:  UART1_BASE,
		Clock: GetUARTClock,
	}

	// Serial port 2
	UART2 = &uart.UART{
		Index: 2,
		Base:  UART2_BASE,
		Clock: GetUARTClock,
	}

	// USB controller 1
	USB1 = &usb.USB{
		Index:     1,
		Base:      USB1_BASE,
		CCGR:      CCM_CCGR6,
		CG:        CCGRx_CG0,
		Analog:    USB_ANALOG1_BASE,
		PHY:       USBPHY1_BASE,
		EnablePLL: EnableUSBPLL,
	}

	// USB controller 2
	USB2 = &usb.USB{
		Index:     2,
		Base:      USB2_BASE,
		CCGR:      CCM_CCGR6,
		CG:        CCGRx_CG0,
		Analog:    USB_ANALOG2_BASE,
		PHY:       USBPHY2_BASE,
		EnablePLL: EnableUSBPLL,
	}

	// SD/MMC controller 1
	USDHC1 = &usdhc.USDHC{
		Index:    1,
		Base:     USDHC1_BASE,
		CCGR:     CCM_CCGR6,
		CG:       CCGRx_CG1,
		SetClock: SetUSDHCClock,
	}

	// SD/MMC controller 2
	USDHC2 = &usdhc.USDHC{
		Index:    2,
		Base:     USDHC2_BASE,
		CCGR:     CCM_CCGR6,
		CG:       CCGRx_CG2,
		SetClock: SetUSDHCClock,
	}
)

// SiliconVersion returns the SoC silicon version information
// (p3945, 57.4.11 Chip Silicon Version (USB_ANALOG_DIGPROG), IMX6ULLRM).
func SiliconVersion() (sv, family, revMajor, revMinor uint32) {
	sv = reg.Read(USB_ANALOG_DIGPROG)

	family = (sv >> 16) & 0xff
	revMajor = (sv >> 8) & 0xff
	revMinor = sv & 0xff

	return
}

// UniqueID returns the NXP SoC Device Unique 64-bit ID.
func UniqueID() (uid [8]byte) {
	OCOTP.Init()

	cfg0, _ := OCOTP.Read(0, 1)
	cfg1, _ := OCOTP.Read(0, 2)

	binary.LittleEndian.PutUint32(uid[0:4], cfg0)
	binary.LittleEndian.PutUint32(uid[4:8], cfg1)

	return
}

// Model returns the SoC model name.
func Model() (model string) {
	switch Family {
	case IMX6UL:
		model = "i.MX6UL"
	case IMX6ULL:
		model = "i.MX6ULL"
	default:
		model = "unknown"
	}

	return
}

// HAB returns whether the SoC is in Trusted or Secure state (indicating that
// Secure Boot is enabled).
func HAB() bool {
	return SNVS.Available()
}