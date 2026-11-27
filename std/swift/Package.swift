// swift-tools-version: 6.0
import PackageDescription

let package = Package(
    name: "Bstd",
    products: [.library(name: "Bstd", targets: ["Bstd"])],
    targets: [
        .target(name: "Bstd"),
        .testTarget(name: "BstdTests", dependencies: ["Bstd"]),
    ]
)
