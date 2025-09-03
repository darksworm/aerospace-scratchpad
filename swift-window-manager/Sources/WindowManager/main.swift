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
    
    func resizeAndPositionWindow(windowID: Int, widthPercent: Int, heightPercent: Int, position: String = "center") -> Bool {
        guard let windowInfo = findWindowByID(windowID) else {
            print("Error: Window with ID \(windowID) not found")
            return false
        }
        
        // Get the screen that contains this window
        let windowScreen = getScreenForWindow(windowInfo.bounds) ?? NSScreen.main
        guard let targetScreen = windowScreen else {
            print("Error: Could not determine target screen")
            return false
        }
        
        let screenFrame = targetScreen.frame
        let screenVisibleFrame = targetScreen.visibleFrame
        
        // For modern MacBooks with notch, use the safe area below the notch
        // The visibleFrame already accounts for menu bar, dock, and notch
        let safeFrame = screenVisibleFrame
        
        // Debug output to show frame information
        print("Window current bounds: \(windowInfo.bounds)")
        print("Using screen: \(targetScreen.localizedName)")
        print("Full screen frame: \(screenFrame)")
        print("Visible frame (safe area below notch): \(safeFrame)")
        
        // Calculate target size as percentage of the safe area (below notch)
        let targetWidth = safeFrame.width * CGFloat(widthPercent) / 100.0
        let targetHeight = safeFrame.height * CGFloat(heightPercent) / 100.0
        
        // Calculate position based on the position parameter
        let (targetX, targetY) = calculatePosition(
            position: position,
            safeFrame: safeFrame,
            targetWidth: targetWidth,
            targetHeight: targetHeight
        )
        
        print("Target size: \(targetWidth) x \(targetHeight)")
        print("Target position: (\(targetX), \(targetY)) - position: \(position)")
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
        var newPosition = CGPoint(x: targetX, y: targetY)
        var newSize = CGSize(width: targetWidth, height: targetHeight)
        
        let positionValue = AXValueCreate(.cgPoint, &newPosition)!
        let sizeValue = AXValueCreate(.cgSize, &newSize)!
        
        // Try to resize first, then position
        let sizeResult = AXUIElementSetAttributeValue(window, kAXSizeAttribute as CFString, sizeValue)
        let positionResult = AXUIElementSetAttributeValue(window, kAXPositionAttribute as CFString, positionValue)
        
        if sizeResult != AXError.success {
            print("Warning: Failed to set window size (error: \(sizeResult.rawValue))")
            print("Fallback: Attempting to position window without resizing...")
            
            // If resize failed, try to position the window at its current size
            // First get the current size to calculate proper positioning
            var currentSizeRef: CFTypeRef?
            if AXUIElementCopyAttributeValue(window, kAXSizeAttribute as CFString, &currentSizeRef) == AXError.success,
               let sizeRef = currentSizeRef {
                var currentSize = CGSize.zero
                if AXValueGetValue(sizeRef as! AXValue, .cgSize, &currentSize) {
                    // Recalculate position based on current size instead of target size
                    let (fallbackX, fallbackY) = calculatePositionForSize(
                        position: position,
                        safeFrame: safeFrame,
                        windowWidth: currentSize.width,
                        windowHeight: currentSize.height
                    )
                    
                    var fallbackPosition = CGPoint(x: fallbackX, y: fallbackY)
                    let fallbackPositionValue = AXValueCreate(.cgPoint, &fallbackPosition)!
                    let fallbackPositionResult = AXUIElementSetAttributeValue(window, kAXPositionAttribute as CFString, fallbackPositionValue)
                    
                    if fallbackPositionResult == AXError.success {
                        print("Fallback positioning successful at (\(fallbackX), \(fallbackY)) with current size \(currentSize.width)x\(currentSize.height)")
                        return true
                    } else {
                        print("Fallback positioning also failed (error: \(fallbackPositionResult.rawValue))")
                    }
                }
            }
        } else if positionResult != AXError.success {
            print("Warning: Failed to set window position (error: \(positionResult.rawValue))")
        }
        
        if positionResult != AXError.success {
            print("Warning: Failed to set window position (error: \(positionResult.rawValue))")
        }
        
        return positionResult == AXError.success || sizeResult == AXError.success
    }
    
    // Calculate window position based on position parameter
    func calculatePosition(position: String, safeFrame: CGRect, targetWidth: CGFloat, targetHeight: CGFloat) -> (CGFloat, CGFloat) {
        return calculatePositionForSize(position: position, safeFrame: safeFrame, windowWidth: targetWidth, windowHeight: targetHeight)
    }
    
    // Calculate window position for given window size
    func calculatePositionForSize(position: String, safeFrame: CGRect, windowWidth: CGFloat, windowHeight: CGFloat) -> (CGFloat, CGFloat) {
        switch position.lowercased() {
        case "center":
            let centerX = safeFrame.origin.x + (safeFrame.width - windowWidth) / 2
            let centerY = safeFrame.origin.y + (safeFrame.height - windowHeight) / 2
            return (centerX, centerY)
            
        case "top":
            let centerX = safeFrame.origin.x + (safeFrame.width - windowWidth) / 2
            let topY = safeFrame.origin.y + 20 // Small margin from top
            return (centerX, topY)
            
        case "bottom":
            let centerX = safeFrame.origin.x + (safeFrame.width - windowWidth) / 2
            let bottomY = safeFrame.origin.y + safeFrame.height - windowHeight - 20 // Small margin from bottom
            return (centerX, bottomY)
            
        case "left":
            let leftX = safeFrame.origin.x + 20 // Small margin from left
            let centerY = safeFrame.origin.y + (safeFrame.height - windowHeight) / 2
            return (leftX, centerY)
            
        case "right":
            let rightX = safeFrame.origin.x + safeFrame.width - windowWidth - 20 // Small margin from right
            let centerY = safeFrame.origin.y + (safeFrame.height - windowHeight) / 2
            return (rightX, centerY)
            
        default:
            // Default to center for unknown positions
            let centerX = safeFrame.origin.x + (safeFrame.width - windowWidth) / 2
            let centerY = safeFrame.origin.y + (safeFrame.height - windowHeight) / 2
            return (centerX, centerY)
        }
    }
    
    // Get the screen that contains the given window bounds
    func getScreenForWindow(_ windowBounds: CGRect) -> NSScreen? {
        let windowCenter = CGPoint(
            x: windowBounds.origin.x + windowBounds.size.width / 2,
            y: windowBounds.origin.y + windowBounds.size.height / 2
        )
        
        // Find the screen that contains the center of the window
        for screen in NSScreen.screens {
            if screen.frame.contains(windowCenter) {
                print("Found window on screen: \(screen.localizedName)")
                return screen
            }
        }
        
        // If no screen contains the center, find the screen with the most overlap
        var bestScreen: NSScreen?
        var maxOverlapArea: CGFloat = 0
        
        for screen in NSScreen.screens {
            let intersection = windowBounds.intersection(screen.frame)
            let overlapArea = intersection.size.width * intersection.size.height
            
            if overlapArea > maxOverlapArea {
                maxOverlapArea = overlapArea
                bestScreen = screen
            }
        }
        
        if let screen = bestScreen {
            print("Using screen with best overlap: \(screen.localizedName)")
        } else {
            print("No suitable screen found, will use main screen")
        }
        
        return bestScreen
    }
}

// Command line interface
func printUsage() {
    print("Usage: window-manager resize <window-id> <width-percent> <height-percent> [position]")
    print("Position options: center (default), top, bottom, left, right")
    print("Examples:")
    print("  window-manager resize 12345 60 90")
    print("  window-manager resize 12345 60 90 bottom")
    print("  window-manager resize 12345 80 70 top")
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
        guard args.count >= 5 && args.count <= 6 else {
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
        
        // Get position parameter if provided, default to "center"
        let position = args.count == 6 ? args[5] : "center"
        
        let windowManager = WindowManager()
        let success = windowManager.resizeAndPositionWindow(
            windowID: windowID,
            widthPercent: widthPercent,
            heightPercent: heightPercent,
            position: position
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