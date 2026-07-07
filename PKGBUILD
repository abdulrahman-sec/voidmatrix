# Maintainer: abdulrahman-sec <abdulrahman-sec@proton.me>
pkgname=voidmatrix-git
pkgver=r0.d910dc7
pkgrel=1
pkgdesc="The ultimate, high-performance Matrix-style digital rain animation simulator for terminal"
arch=('x86_64' 'aarch64')
url="https://github.com/abdulrahman-sec/voidmatrix"
license=('MIT')
depends=('glibc')
makedepends=('go' 'git')
provides=('voidmatrix')
conflicts=('voidmatrix')
source=("git+https://github.com/abdulrahman-sec/voidmatrix.git")
sha256sums=('SKIP')

pkgver() {
  cd "$srcdir/${pkgname%-git}"
  printf "r%s.%s" "$(git rev-list --count HEAD)" "$(git rev-parse --short HEAD)"
}

prepare() {
  cd "$srcdir/${pkgname%-git}"
  mkdir -p build
}

build() {
  cd "$srcdir/${pkgname%-git}"
  export GOPATH="$srcdir/gopath"
  go build \
    -trimpath \
    -buildmode=pie \
    -mod=readonly \
    -modcacherw \
    -ldflags "-linkmode external -extldflags \"${LDFLAGS}\"" \
    -o build/voidmatrix .
}

package() {
  cd "$srcdir/${pkgname%-git}"
  install -Dm755 build/voidmatrix "$pkgdir/usr/bin/voidmatrix"
  install -Dm644 README.md "$pkgdir/usr/share/doc/voidmatrix/README.md"
  install -Dm644 LICENSE "$pkgdir/usr/share/licenses/voidmatrix/LICENSE"
}
