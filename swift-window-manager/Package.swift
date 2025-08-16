// swift-tools-version:5.5
import PackageDescription

let package = Package(
    name: "WindowManager",
    platforms: [
        .macOS(.v10_15)
    ],
    products: [
        .executable(name: "window-manager", targets: ["WindowManager"])
    ],
    targets: [
        .executableTarget(
            name: "WindowManager",
            dependencies: []
        )
    ]
)