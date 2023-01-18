# appimage-runtime

New runtime for AppImages that uses both fuse3 (preferred) or fuse2.

As of right now, this is considered alpha quality and needs significantly more testing. Somewhat based off of <https://github.com/orivej/static-appimage/>

## Testing

You can build the runtime, as well as attach the runtime to existing AppImages using `build.sh`. You can additionally use `attach/attach.go` (which is made to be run using `go run`) on it's own to attach the runtime to existing AppImages.

AppImage's using this runtime might now integrate properly with some applications (namely appimaged from `https://github.com/probonopd/go-appimage` would not be able to integrate it properly).
