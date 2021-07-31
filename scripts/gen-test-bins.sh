TESTDIRS="./..."
pkg_list=$(go list $TESTDIRS | grep -vE "($exclude_paths)")

for pkg in $pkg_list
do
    echo "generating test bin for $pkg"
    name=$(echo "$pkg" | sed "s/\//-/g")
    go test -c -o $1/$name $pkg
done
