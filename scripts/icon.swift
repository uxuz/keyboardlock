// Draws the app icon to the template of macOS: a 1024×1024 canvas holding
// an 824×824 rounded square with continuous corners and its shadow.
//
//	swift scripts/icon.swift resources/icon.png
//
// The glyph is "keyboard-off" from Tabler Icons by Paweł Kuna, under the
// MIT License (https://github.com/tabler/tabler-icons/blob/master/LICENSE).
import AppKit
import SwiftUI

let size = 1024
let space = CGColorSpace(name: CGColorSpace.sRGB)!
let ctx = CGContext(data: nil, width: size, height: size, bitsPerComponent: 8, bytesPerRow: 0,
                    space: space, bitmapInfo: CGImageAlphaInfo.premultipliedLast.rawValue)!
// The origin at the top left, as in the SVG of the glyph.
ctx.translateBy(x: 0, y: CGFloat(size))
ctx.scaleBy(x: 1, y: -1)

func color(_ hex: UInt32, _ alpha: CGFloat = 1) -> CGColor {
    CGColor(colorSpace: space, components: [CGFloat(hex >> 16 & 0xFF) / 255, CGFloat(hex >> 8 & 0xFF) / 255, CGFloat(hex & 0xFF) / 255, alpha])!
}

// The tile and its shadow, whose offset is in pixels of the bitmap.
let tile = RoundedRectangle(cornerRadius: 185.4, style: .continuous)
    .path(in: CGRect(x: 100, y: 100, width: 824, height: 824)).cgPath
ctx.saveGState()
ctx.setShadow(offset: CGSize(width: 0, height: -12), blur: 28, color: color(0x000000, 0.3))
ctx.addPath(tile)
ctx.setFillColor(color(0x1F8BFF))
ctx.fillPath()
ctx.restoreGState()

ctx.saveGState()
ctx.addPath(tile)
ctx.clip()
let gradient = CGGradient(colorsSpace: space, colors: [color(0x55ABFF), color(0x0072F5)] as CFArray, locations: [0, 1])!
ctx.drawLinearGradient(gradient, start: CGPoint(x: 0, y: 100), end: CGPoint(x: 0, y: 924), options: [])
ctx.restoreGState()

// The glyph, in the 24 units of its SVG.
let glyph = CGMutablePath()
glyph.move(to: CGPoint(x: 18, y: 18))
glyph.addArc(tangent1End: CGPoint(x: 2, y: 18), tangent2End: CGPoint(x: 2, y: 8), radius: 2)
glyph.addArc(tangent1End: CGPoint(x: 2, y: 6), tangent2End: CGPoint(x: 6, y: 6), radius: 2)
glyph.addLine(to: CGPoint(x: 6, y: 6))
glyph.move(to: CGPoint(x: 10, y: 6))
glyph.addArc(tangent1End: CGPoint(x: 22, y: 6), tangent2End: CGPoint(x: 22, y: 16), radius: 2)
glyph.addLine(to: CGPoint(x: 22, y: 16))
glyph.addCurve(to: CGPoint(x: 21.41, y: 17.418), control1: CGPoint(x: 22, y: 16.554), control2: CGPoint(x: 21.774, y: 17.056))
for (x, y) in [(6, 10), (10, 10), (14, 10), (18, 10), (6, 14), (18, 14)] {
    glyph.move(to: CGPoint(x: Double(x), y: Double(y)))
    glyph.addLine(to: CGPoint(x: Double(x), y: Double(y) + 0.01))
}
glyph.move(to: CGPoint(x: 10, y: 14))
glyph.addLine(to: CGPoint(x: 14, y: 14))
glyph.move(to: CGPoint(x: 3, y: 3))
glyph.addLine(to: CGPoint(x: 21, y: 21))

let scale: CGFloat = 24
ctx.saveGState()
ctx.setShadow(offset: CGSize(width: 0, height: -6), blur: 14, color: color(0x003C8F, 0.35))
ctx.translateBy(x: 512 - 12 * scale, y: 512 - 12 * scale)
ctx.scaleBy(x: scale, y: scale)
ctx.addPath(glyph)
ctx.setStrokeColor(color(0xFFFFFF))
ctx.setLineWidth(2)
ctx.setLineCap(.round)
ctx.setLineJoin(.round)
ctx.strokePath()
ctx.restoreGState()

let png = NSBitmapImageRep(cgImage: ctx.makeImage()!).representation(using: .png, properties: [:])!
try png.write(to: URL(fileURLWithPath: CommandLine.arguments.count > 1 ? CommandLine.arguments[1] : "icon.png"))
