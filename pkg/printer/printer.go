package printer

import (
	"fmt"
	"net"
	"os"
	"time"
)

// Printer is the interface for sending raw ESC/POS data to a thermal printer.
type Printer interface {
	// Print sends raw ESC/POS bytes to the printer.
	Print(data []byte) error
	// Close releases the printer connection/handle.
	Close() error
	// IsConnected returns true if the printer connection is active.
	IsConnected() bool
}

// --- USB Printer (writes to device file, e.g. /dev/usb/lp0) ---

type usbPrinter struct {
	path string
}

// NewUSBPrinter creates a printer that writes to a USB device file.
func NewUSBPrinter(devicePath string) Printer {
	return &usbPrinter{path: devicePath}
}

func (p *usbPrinter) Print(data []byte) error {
	f, err := os.OpenFile(p.path, os.O_WRONLY, 0)
	if err != nil {
		return fmt.Errorf("printer: failed to open USB device %s: %w", p.path, err)
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("printer: failed to write to USB device %s: %w", p.path, err)
	}
	return nil
}

func (p *usbPrinter) Close() error {
	return nil // USB printer opens/closes per print job
}

func (p *usbPrinter) IsConnected() bool {
	_, err := os.Stat(p.path)
	return err == nil
}

// --- Network Printer (dials TCP, e.g. 192.168.1.100:9100) ---

type networkPrinter struct {
	address string
	timeout time.Duration
}

// NewNetworkPrinter creates a printer that connects via TCP.
// Address should include port, e.g. "192.168.1.100:9100".
func NewNetworkPrinter(address string) Printer {
	return &networkPrinter{
		address: address,
		timeout: 5 * time.Second,
	}
}

func (p *networkPrinter) Print(data []byte) error {
	conn, err := net.DialTimeout("tcp", p.address, p.timeout)
	if err != nil {
		return fmt.Errorf("printer: failed to connect to %s: %w", p.address, err)
	}
	defer conn.Close()

	_ = conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("printer: failed to write to %s: %w", p.address, err)
	}
	return nil
}

func (p *networkPrinter) Close() error {
	return nil // Network printer opens/closes per print job
}

func (p *networkPrinter) IsConnected() bool {
	conn, err := net.DialTimeout("tcp", p.address, 2*time.Second)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// --- Null Printer (no-op, used when no printer is configured) ---

type nullPrinter struct{}

// NewNullPrinter creates a no-op printer for environments without hardware.
func NewNullPrinter() Printer {
	return &nullPrinter{}
}

func (p *nullPrinter) Print(data []byte) error {
	return nil
}

func (p *nullPrinter) Close() error {
	return nil
}

func (p *nullPrinter) IsConnected() bool {
	return false
}

// NewPrinterFromConfig creates the appropriate Printer based on type.
//
//	printerType: "usb", "network", or "none"
//	usbPath: device path for USB printers (e.g. "/dev/usb/lp0")
//	address: TCP address for network printers (e.g. "192.168.1.100:9100")
func NewPrinterFromConfig(printerType, usbPath, address string) (Printer, error) {
	switch printerType {
	case "usb":
		if usbPath == "" {
			return nil, fmt.Errorf("printer: USB path is required for USB printer type")
		}
		return NewUSBPrinter(usbPath), nil
	case "network":
		if address == "" {
			return nil, fmt.Errorf("printer: address is required for network printer type")
		}
		return NewNetworkPrinter(address), nil
	case "none", "":
		return NewNullPrinter(), nil
	default:
		return nil, fmt.Errorf("printer: unknown printer type %q (use usb, network, or none)", printerType)
	}
}
