#!/bin/bash

if [ -n "$1" ]; then
  version="$1"
elif [ -n "$VER" ]; then
  version="$VER"
else
  version="dev"
fi
name='ovi'
# version=${VER:-dev} #'v1.1.2'
tmp=$name-$version
bin='bin/'
golang='go' # go1.21.13

if [ -d $bin$version ]; then
    rm -r $bin$version
fi
mkdir -p $bin$version
bin+=$version/

compress(){
    echo - $out
    if [ ! -f $name$ext ]; then
        echo build failed!
        exit 1
    fi
    zip -9 $bin$out.zip $name$ext
    # mv $name$ext $bin$out$ext
    rm $name$ext
}

tagdef='' # go_json nomsgpack
tagsonic='' # sonic avx nomsgpack
ldfext='-s -w -linkmode external'
ldfstatic="$ldfext -extldflags '-static'"


echo 'Build: linux-amd64'

export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=1
export CC=x86_64-linux-gnu-gcc
export CXX=x86_64-linux-gnu-g++
export AR=x86_64-linux-gnu-gcc-ar

base=$tmp-linux-amd64

export GOAMD64=v1
out=${base}v1-gnu
$golang build -trimpath -tags "$tagdef" -o $name -ldflags "$ldfext"
compress

export GOAMD64=v2
out=${base}v2-gnu
$golang build -trimpath -tags "$tagdef" -o $name -ldflags "$ldfext"
compress

export GOAMD64=v3
out=${base}v3-gnu
$golang build -trimpath -tags "$tagsonic" -o $name -ldflags "$ldfext"
compress

# tc=/usr/local/x86_64-linux-musl-native/bin

# export CC=$tc/x86_64-linux-musl-gcc
# export CXX=$tc/x86_64-linux-musl-c++
# export AR=$tc/x86_64-linux-musl-gcc-ar

# export GOAMD64=v1
# out=${base}v1-musl
# go1.21.13 build -trimpath -tags "$tagdef" -o $name -ldflags "$ldfext"
# compress

# out+='-static'
# go1.21.13 build -trimpath -tags "$tagdef" -o $name -ldflags "$ldfstatic"
# compress

# export GOAMD64=v2
# out=${base}v2-musl
# go1.21.13 build -trimpath -tags "$tagdef" -o $name -ldflags "$ldfext"
# compress

# out+='-static'
# go1.21.13 build -trimpath -tags "$tagdef" -o $name -ldflags "$ldfstatic"
# compress

# export GOAMD64=v3
# out=${base}v3-musl
# go1.21.13 build -trimpath -tags "$tagsonic" -o $name -ldflags "$ldfext"
# compress

# out+='-static'
# go1.21.13 build -trimpath -tags "$tagsonic" -o $name -ldflags "$ldfstatic"
# compress


echo 'Build: linux-arm7'

base=$tmp-linux-arm7

export GOOS=linux
export GOARCH=arm
export GOARM=7
export CGO_ENABLED=1

out=$base-gnu

export CC=arm-linux-gnueabihf-gcc
export CXX=arm-linux-gnueabihf-cpp
export AR=arm-linux-gnueabihf-gcc-ar

$golang build -trimpath -tags "$tagdef" -o $name -ldflags "$ldfext"
compress

out=$base-musl
tc=/usr/local/armv7l-linux-musleabihf-cross/bin

export CC=$tc/armv7l-linux-musleabihf-gcc
export CXX=$tc/armv7l-linux-musleabihf-c++
export AR=$tc/armv7l-linux-musleabihf-gcc-ar

$golang build -trimpath -tags "$tagdef" -o $name -ldflags "$ldfext"
compress

out+='-static'
$golang build -trimpath -tags "$tagdef" -o $name -ldflags "$ldfstatic"
compress


echo 'Build: linux-arm64'

base=$tmp-linux-arm64

export GOARCH=arm64

out=$base-gnu

export CC=aarch64-linux-gnu-gcc
export CXX=aarch64-linux-gnu-cpp
export AR=aarch64-linux-gnu-gcc-ar

$golang build -trimpath -tags "$tagdef" -o $name -ldflags "$ldfext"
compress

out=$base-musl
tc=/usr/local/aarch64-linux-musl-cross/bin

export CC=$tc/aarch64-linux-musl-gcc
export CXX=$tc/aarch64-linux-musl-g++
export AR=$tc/aarch64-linux-musl-gcc-ar

$golang build -trimpath -tags "$tagdef" -o $name -ldflags "$ldfext"
compress

out+='-static'
$golang build -trimpath -tags "$tagdef" -o $name -ldflags "$ldfstatic"
compress


echo 'Build: freebsd-amd64'

export GOOS=freebsd
export GOARCH=amd64
export CGO_ENABLED=0

base=$tmp-freebsd-amd64

export GOAMD64=v2
out=${base}v2-pure
$golang build -trimpath -tags "$tagdef" -o $name -ldflags "-s -w"
compress

# export GOAMD64=v3
# out=${base}v3-pure
# go1.21.13 build -trimpath -tags "$tagsonic" -o $name -ldflags "-s -w"
# compress


echo 'Build: android'

export GOOS=android
export CGO_ENABLED=1

base=$tmp-android-
tc=/usr/local/android-ndk-r26b/toolchains/llvm/prebuilt/linux-x86_64/bin
export AR=$tc/llvm-ar

# export GOARCH=amd64
# export GOAMD64=v2
# export CC=$tc/x86_64-linux-android24-clang
# export CXX=$tc/x86_64-linux-android24-clang++
# out=${base}amd64v2-ndk24
# go1.21.13 build -trimpath -tags "$tagdef" -o $name -ldflags "$ldfext"
# compress

export GOARCH=arm64
export CC=$tc/aarch64-linux-android24-clang
export CXX=$tc/aarch64-linux-android24-clang++
out=${base}arm64-ndk24
$golang build -trimpath -tags "$tagdef" -o $name -ldflags "$ldfext"
compress

# export GOARCH=arm
# export GOARM=7
# export CC=$tc/armv7a-linux-androideabi24-clang
# export CXX=$tc/armv7a-linux-androideabi24-clang++
# out=${base}arm7-ndk24
# go1.21.13 build -trimpath -tags "$tagdef" -o $name -ldflags "$ldfext"
# compress


echo 'Build: windows-amd64'

export GOOS=windows
export GOARCH=amd64
export CGO_ENABLED=1
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-cpp
export AR=x86_64-w64-mingw32-gcc-ar

base=$tmp-windows-amd64
ext='.exe'

export GOAMD64=v2
out=${base}v2-mingw32
$golang build -trimpath -tags "$tagdef" -o $name$ext -ldflags "$ldfext"
compress

export GOAMD64=v3
out=${base}v3-mingw32
$golang build -trimpath -tags "$tagsonic" -o $name$ext -ldflags "$ldfext"
compress


# echo 'Build: win7-amd64'

# tc=/usr/local/x86_64-w64-mingw32-cross/bin

# export CC=x86_64-w64-mingw32-gcc
# export CXX=x86_64-w64-mingw32-cpp
# export AR=x86_64-w64-mingw32-gcc-ar

# base=$tmp-win7-amd64

# export GOAMD64=v1
# out=${base}v1-mingw32
# go1.21.13 build -trimpath -tags "$tagdef" -o $name$ext -ldflags "-s -w"
# compress

# export GOAMD64=v2
# out=${base}v2-mingw32
# go1.20.14 build -trimpath -tags "$tagdef" -o $name$ext -ldflags "-s -w"
# compress

# export GOAMD64=v3
# out=${base}v3-mingw32
# go1.20.14 build -trimpath -tags "$tagsonic" -o $name$ext -ldflags "-s -w"
# compress


echo 'Done!'