#!/bin/sh

# Largely based off of https://github.com/probonopd/go-appimage/blob/master/scripts/build.sh

# build GOARCH
build () {
    case $1 in
        amd64) export ZIGTARGET=x86_64-linux-musl;;
        386) export ZIGTARGET=i386-linux-musl;;
        arm64) export ZIGTARGET=aarch64-linux-musl;;
        arm) export ZIGTARGET=arm-linux-musleabihf;;
    esac
    export CC="zig cc -Wl,--strip-all -target $ZIGTARGET"
    GOARCH=$1 CGO_ENABLED=1 go build -o $BUILDDIR/static-appimage-$1 -trimpath -ldflags="-linkmode=external -s -w"
}

help_message() {
  echo "build.sh -a [arch] -p"
  echo ""
  echo "  -a"
  echo "    Comma seperated list of architectures to build for (as defined by GOARCH)."
  echo "    If not given, only the host architecture is built. Accepts amd64, 386, arm64, and arm"
  echo "  -p"
  echo "    After building the runtime process all AppImages in this directory, replacing their runtime with this runtime."
  echo "    Uses attach/attach.go"
  echo "  -s"
  echo "    Used with -p. Strip the AppImage and save it before attaching the new runtime"
  echo "  -h"
  echo "    Print this message"
  exit 0
}

while [ $# -gt 0 ]; do
  case $1 in
    -p)
      ATTACH=true;;
    -a)
      BUILDARCH=(${2//,/ })
      shift;;
    -s)
      STRIP=true;;
    -h)
      help_message;;
    *)
      echo "Invalid parameter $1"
      exit 1;;
  esac
  shift
done

BUILDDIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )/build

mkdir -p $BUILDDIR || true

if [ ! -e $BUILDDIR/zig ]; then
  wget -c -q "https://ziglang.org/download/0.10.0/zig-linux-x86_64-0.10.0.tar.xz"
  tar xf zig-linux-*.tar.xz
  rm zig-linux-*.tar.xz
  mv zig-linux-* $BUILDDIR/zig
fi

PATH=$BUILDDIR/zig:$PATH

if [ -z $BUILDARCH ]; then
  BUILDARCH=(386 amd64 arm arm64)
fi

for arch in ${BUILDARCH[@]}; do
  build $arch
done

go run stamp/stamp.go  $BUILDDIR/static-appimage-*

if [ ! -z $ATTACH ]; then
  HOSTARCH=$(go env GOHOSTARCH)
  if [ -z $STRIP ]; then
    go run attach/attach.go -r $BUILDDIR/static-appimage-$HOSTARCH
  else
    go run attach/attach.go -s -r $BUILDDIR/static-appimage-$HOSTARCH
  fi
fi

if [ -z $DONTCLEAN ]; then
  for file in ${CLEANUP[@]}; do
    echo $file
    rm -rf $file || true
  done
fi