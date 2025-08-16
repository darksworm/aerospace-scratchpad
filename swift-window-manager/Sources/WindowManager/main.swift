import Foundation
import ApplicationServices
import CoreGraphics
import AppKit

struct WindowInfo {
    let windowID: Int
    let pid: pid_t
    let appName: String
    let bounds: CGRect
}

class WindowManager {
    
    init() {
        // Check if we have accessibility permissions
        if !AXIsProcessTrusted() {
            print("Error: Accessibility permissions required. Please grant permissions in System Preferences > Security & Privacy > Privacy > Accessibility")
            exit(1)
        }
    }
    
    func findWindowByID(_ windowID: Int) -> WindowInfo? {
        guard let windowList = CGWindowListCopyWindowInfo([.optionOnScreenOnly], kCGNullWindowID) as? [[String: Any]] else {
            return nil
        }
        
        for windowData in windowList {
            if let currentWindowID = windowData[kCGWindowNumber as String] as? Int,
               currentWindowID == windowID,
               let pid = windowData[kCGWindowOwnerPID as String] as? pid_t,
               let appName = windowData[kCGWindowOwnerName as String] as? String,
               let boundsDict = windowData[kCGWindowBounds as String] as? [String: Any] {
                
                let bounds = CGRect(
                    x: boundsDict["X"] as? CGFloat ?? 0,
                    y: boundsDict["Y"] as? CGFloat ?? 0,
                    width: boundsDict["Width"] as? CGFloat ?? 0,
                    height: boundsDict["Height"] as? CGFloat ?? 0
                )
                
                return WindowInfo(windowID: windowID, pid: pid, appName: appName, bounds: bounds)
            }
        }
        
        return nil
    }
    
    func resizeAndPositionWindow(windowID: Int, widthPercent: Int, heightPercent: Int) -> Bool {
        guard let windowInfo = findWindowByID(windowID) else {
            print("Error: Window with ID \(windowID) not found")
            return false
        }
        
        // Get main screen dimensions
        guard let mainScreen = NSScreen.main else {
            print("Error: Could not get main screen")
            return false
        }
        
        let screenFrame = mainScreen.frame
        let screenVisibleFrame = mainScreen.visibleFrame
        
        // For modern MacBooks with notch, use the safe area below the notch
        // The visibleFrame already accounts for menu bar, dock, and notch
        let safeFrame = screenVisibleFrame
        
        // Debug output to show frame information
        print("Full screen frame: \(screenFrame)")
        print("Visible frame (safe area below notch): \(safeFrame)")
        
        // Calculate target size as percentage of the safe area (below notch)
        let targetWidth = safeFrame.width * CGFloat(widthPercent) / 100.0
        let targetHeight = safeFrame.height * CGFloat(heightPercent) / 100.0
        
        // Calculate centered position within the safe area (below notch)
        // This centers the window in the available space, not the full screen
        let centerX = safeFrame.origin.x + (safeFrame.width - targetWidth) / 2
        let centerY = safeFrame.origin.y + (safeFrame.height - targetHeight) / 2
        
        print("Target size: \(targetWidth) x \(targetHeight)")
        print("Target position: (\(centerX), \(centerY)) - centered in safe area")
        print("Safe area dimensions: \(safeFrame.width) x \(safeFrame.height)")
        print("Safe area origin: (\(safeFrame.origin.x), \(safeFrame.origin.y))")
        
        // Create application reference
        let appRef = AXUIElementCreateApplication(windowInfo.pid)
        
        // Get windows for this application
        var windowsRef: CFTypeRef?
        let result = AXUIElementCopyAttributeValue(appRef, kAXWindowsAttribute as CFString, &windowsRef)
        
        guard result == AXError.success, let windows = windowsRef as? [AXUIElement] else {
            print("Error: Could not get windows for application")
            return false
        }
        
        // Find the specific window by comparing bounds (since we can't directly match by window ID)
        var targetWindow: AXUIElement?
        
        for window in windows {
            var positionRef: CFTypeRef?
            var sizeRef: CFTypeRef?
            
            if AXUIElementCopyAttributeValue(window, kAXPositionAttribute as CFString, &positionRef) == AXError.success,
               AXUIElementCopyAttributeValue(window, kAXSizeAttribute as CFString, &sizeRef) == AXError.success {
                
                var currentPosition = CGPoint.zero
                var currentSize = CGSize.zero
                
                if AXValueGetValue(positionRef as! AXValue, .cgPoint, &currentPosition),
                   AXValueGetValue(sizeRef as! AXValue, .cgSize, &currentSize) {
                    
                    let currentBounds = CGRect(origin: currentPosition, size: currentSize)
                    
                    // Match by bounds (with some tolerance for floating point precision)
                    if abs(currentBounds.origin.x - windowInfo.bounds.origin.x) < 1.0 &&
                       abs(currentBounds.origin.y - windowInfo.bounds.origin.y) < 1.0 &&
                       abs(currentBounds.size.width - windowInfo.bounds.size.width) < 1.0 &&
                       abs(currentBounds.size.height - windowInfo.bounds.size.height) < 1.0 {
                        targetWindow = window
                        break
                    }
                }
            }
        }
        
        guard let window = targetWindow else {
            print("Error: Could not find target window in application")
            return false
        }
        
        // Set new position and size
        var newPosition = CGPoint(x: centerX, y: centerY)
        var newSize = CGSize(width: targetWidth, height: targetHeight)
        
        let positionValue = AXValueCreate(.cgPoint, &newPosition)!
        let sizeValue = AXValueCreate(.cgSize, &newSize)!
        
        let positionResult = AXUIElementSetAttributeValue(window, kAXPositionAttribute as CFString, positionValue)
        let sizeResult = AXUIElementSetAttributeValue(window, kAXSizeAttribute as CFString, sizeValue)
        
        if positionResult != AXError.success {
            print("Warning: Failed to set window position (error: \(positionResult.rawValue))")
        }
        
        if sizeResult != AXError.success {
            print("Warning: Failed to set window size (error: \(sizeResult.rawValue))")
        }
        
        return positionResult == AXError.success && sizeResult == AXError.success
    }
}

// Command line interface
func printUsage() {
    print("Usage: window-manager resize <window-id> <width-percent> <height-percent>")
    print("Example: window-manager resize 12345 60 90")
}

func main() {
    let args = CommandLine.arguments
    
    guard args.count >= 2 else {
        printUsage()
        exit(1)
    }
    
    let command = args[1]
    
    switch command {
    case "resize":
        guard args.count == 5 else {
            printUsage()
            exit(1)
        }
        
        guard let windowID = Int(args[2]),
              let widthPercent = Int(args[3]),
              let heightPercent = Int(args[4]) else {
            print("Error: Invalid arguments. Window ID, width percent, and height percent must be integers.")
            exit(1)
        }
        
        guard widthPercent > 0 && widthPercent <= 100,
              heightPercent > 0 && heightPercent <= 100 else {
            print("Error: Width and height percentages must be between 1 and 100")
            exit(1)
        }
        
        let windowManager = WindowManager()
        let success = windowManager.resizeAndPositionWindow(
            windowID: windowID,
            widthPercent: widthPercent,
            heightPercent: heightPercent
        )
        
        if success {
            print("Successfully resized and positioned window \(windowID)")
            exit(0)
        } else {
            print("Failed to resize and position window \(windowID)")
            exit(1)
        }
        
    default:
        print("Error: Unknown command '\(command)'")
        printUsage()
        exit(1)
    }
}

main()