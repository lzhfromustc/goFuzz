TESTDIRS="./..."
pkg_list=$(go list $TESTDIRS | grep -vE "($exclude_paths)")

for pkg in $pkg_list
do
    name=$(echo "$pkg" | sed "s/\//-/g")
    go test -c -o testbins/$name $pkg
done
