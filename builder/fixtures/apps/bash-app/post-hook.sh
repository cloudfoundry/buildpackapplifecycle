# test if variables exported by pre-hook.sh are visible
if [ -n "$PREHOOK" ]; then
    # test that always-detect/bin/compile has executed (and has seen $PREHOOK)
    if [ -f PREHOOK ]; then
        touch POSTHOOK
    fi
fi
