package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics -framework Foundation

#import <CoreGraphics/CoreGraphics.h>
#import <Foundation/Foundation.h>

typedef struct {
    int pid;
    char title[256];
    int x;
    int y;
    int width;
    int height;
} WindowInfo;

WindowInfo* getWindowList(int* count) {
    CFArrayRef windowList = CGWindowListCopyWindowInfo(kCGWindowListOptionOnScreenOnly | kCGWindowListExcludeDesktopElements, kCGNullWindowID);
    CFIndex windowCount = CFArrayGetCount(windowList);

    WindowInfo* windows = (WindowInfo*)malloc(sizeof(WindowInfo) * windowCount);
    *count = (int)windowCount;

    for (CFIndex i = 0; i < windowCount; i++) {
        CFDictionaryRef windowInfo = (CFDictionaryRef)CFArrayGetValueAtIndex(windowList, i);

        CFNumberRef pidRef = (CFNumberRef)CFDictionaryGetValue(windowInfo, kCGWindowOwnerPID);
        if (pidRef) {
            CFNumberGetValue(pidRef, kCFNumberIntType, &windows[i].pid);
        }

        CFStringRef titleRef = (CFStringRef)CFDictionaryGetValue(windowInfo, kCGWindowName);
        if (titleRef) {
            CFStringGetCString(titleRef, windows[i].title, 256, kCFStringEncodingUTF8);
        }

        CFDictionaryRef boundsRef = (CFDictionaryRef)CFDictionaryGetValue(windowInfo, kCGWindowBounds);
        if (boundsRef) {
            CFNumberRef xRef = (CFNumberRef)CFDictionaryGetValue(boundsRef, CFSTR("X"));
            CFNumberRef yRef = (CFNumberRef)CFDictionaryGetValue(boundsRef, CFSTR("Y"));
            CFNumberRef widthRef = (CFNumberRef)CFDictionaryGetValue(boundsRef, CFSTR("Width"));
            CFNumberRef heightRef = (CFNumberRef)CFDictionaryGetValue(boundsRef, CFSTR("Height"));

            if (xRef) CFNumberGetValue(xRef, kCFNumberIntType, &windows[i].x);
            if (yRef) CFNumberGetValue(yRef, kCFNumberIntType, &windows[i].y);
            if (widthRef) CFNumberGetValue(widthRef, kCFNumberIntType, &windows[i].width);
            if (heightRef) CFNumberGetValue(heightRef, kCFNumberIntType, &windows[i].height);
        }
    }

    CFRelease(windowList);
    return windows;
}

void freeWindowList(WindowInfo* windows) {
    free(windows);
}
*/
import "C"
import (
	"fmt"
	"unsafe"
)

type WindowInfo struct {
	PID    int
	Title  string
	X      int
	Y      int
	Width  int
	Height int
}

func GetAllWindows() ([]WindowInfo, error) {
	var count C.int
	cWindows := C.getWindowList(&count)
	defer C.freeWindowList(cWindows)

	if count == 0 {
		return nil, fmt.Errorf("no windows found")
	}

	windows := make([]WindowInfo, count)
	cWindowArray := (*[1 << 28]C.WindowInfo)(unsafe.Pointer(cWindows))[:count:count]

	for i := 0; i < int(count); i++ {
		windows[i] = WindowInfo{
			PID:    int(cWindowArray[i].pid),
			Title:  C.GoString(&cWindowArray[i].title[0]),
			X:      int(cWindowArray[i].x),
			Y:      int(cWindowArray[i].y),
			Width:  int(cWindowArray[i].width),
			Height: int(cWindowArray[i].height),
		}
	}

	return windows, nil
}
