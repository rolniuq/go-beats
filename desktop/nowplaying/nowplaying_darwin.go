//go:build desktop

package nowplaying

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework MediaPlayer
#import <MediaPlayer/MediaPlayer.h>
#import <Cocoa/Cocoa.h>

extern void goNowPlayingCommand(int command, double value);

void nowplaying_setPlaying(const char* title, double duration) {
    MPNowPlayingInfoCenter* center = [MPNowPlayingInfoCenter defaultCenter];
    NSString* titleStr = [NSString stringWithUTF8String:title];
    NSImage* appIcon = [NSImage imageNamed:@"AppIcon"];
    if (appIcon == nil) {
        appIcon = [NSImage imageNamed:@"NSApplicationIcon"];
    }
    MPMediaItemArtwork* artwork = nil;
    if (appIcon) {
        artwork = [[MPMediaItemArtwork alloc] initWithBoundsSize:appIcon.size
                                                   requestHandler:^NSImage* (CGSize size) {
            return appIcon;
        }];
    }
    NSMutableDictionary* info = [NSMutableDictionary dictionary];
    info[MPMediaItemPropertyTitle] = titleStr;
    info[MPMediaItemPropertyArtist] = @"go-beats";
    info[MPMediaItemPropertyAlbumTitle] = @"Lofi Beats";
    info[MPNowPlayingInfoPropertyPlaybackRate] = @(1.0);
    if (duration > 0) {
        info[MPMediaItemPropertyPlaybackDuration] = @(duration);
    }
    if (artwork) {
        info[MPMediaItemPropertyArtwork] = artwork;
    }
    center.nowPlayingInfo = info;
}

void nowplaying_setStation(const char* name) {
    MPNowPlayingInfoCenter* center = [MPNowPlayingInfoCenter defaultCenter];
    NSString* nameStr = [NSString stringWithUTF8String:name];
    NSImage* appIcon = [NSImage imageNamed:@"AppIcon"];
    if (appIcon == nil) {
        appIcon = [NSImage imageNamed:@"NSApplicationIcon"];
    }
    MPMediaItemArtwork* artwork = nil;
    if (appIcon) {
        artwork = [[MPMediaItemArtwork alloc] initWithBoundsSize:appIcon.size
                                                   requestHandler:^NSImage* (CGSize size) {
            return appIcon;
        }];
    }
    NSMutableDictionary* info = [NSMutableDictionary dictionary];
    info[MPMediaItemPropertyTitle] = nameStr;
    info[MPMediaItemPropertyArtist] = @"go-beats";
    info[MPMediaItemPropertyAlbumTitle] = @"Radio";
    info[MPNowPlayingInfoPropertyPlaybackRate] = @(1.0);
    info[MPMediaItemPropertyPlaybackDuration] = @(-1);
    if (artwork) {
        info[MPMediaItemPropertyArtwork] = artwork;
    }
    center.nowPlayingInfo = info;
}

void nowplaying_setPaused() {
    MPNowPlayingInfoCenter* center = [MPNowPlayingInfoCenter defaultCenter];
    NSMutableDictionary* info = [center.nowPlayingInfo mutableCopy];
    if (info) {
        info[MPNowPlayingInfoPropertyPlaybackRate] = @(0.0);
        center.nowPlayingInfo = info;
    }
}

void nowplaying_setStopped() {
    [MPNowPlayingInfoCenter defaultCenter].nowPlayingInfo = nil;
}

void nowplaying_setProgress(double position) {
    MPNowPlayingInfoCenter* center = [MPNowPlayingInfoCenter defaultCenter];
    NSMutableDictionary* info = [center.nowPlayingInfo mutableCopy];
    if (info) {
        info[MPNowPlayingInfoPropertyElapsedPlaybackTime] = @(position);
        center.nowPlayingInfo = info;
    }
}

static void handleCommand(int command, double value) {
    goNowPlayingCommand(command, value);
}

void nowplaying_setCommandHandler() {
    MPRemoteCommandCenter* cmdCenter = [MPRemoteCommandCenter sharedCommandCenter];
    [cmdCenter.playCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent* _Nonnull event) {
        handleCommand(0, 0);
        return MPRemoteCommandHandlerStatusSuccess;
    }];
    [cmdCenter.pauseCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent* _Nonnull event) {
        handleCommand(1, 0);
        return MPRemoteCommandHandlerStatusSuccess;
    }];
    [cmdCenter.togglePlayPauseCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent* _Nonnull event) {
        handleCommand(2, 0);
        return MPRemoteCommandHandlerStatusSuccess;
    }];
    [cmdCenter.nextTrackCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent* _Nonnull event) {
        handleCommand(3, 0);
        return MPRemoteCommandHandlerStatusSuccess;
    }];
    [cmdCenter.previousTrackCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent* _Nonnull event) {
        handleCommand(4, 0);
        return MPRemoteCommandHandlerStatusSuccess;
    }];
    [cmdCenter.changePlaybackPositionCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent* _Nonnull event) {
        MPChangePlaybackPositionCommandEvent* posEvent = (MPChangePlaybackPositionCommandEvent*)event;
        handleCommand(5, posEvent.positionTime);
        return MPRemoteCommandHandlerStatusSuccess;
    }];
}

void nowplaying_clearCommandHandler() {
    MPRemoteCommandCenter* cmdCenter = [MPRemoteCommandCenter sharedCommandCenter];
    [cmdCenter.playCommand removeTarget:nil];
    [cmdCenter.pauseCommand removeTarget:nil];
    [cmdCenter.togglePlayPauseCommand removeTarget:nil];
    [cmdCenter.nextTrackCommand removeTarget:nil];
    [cmdCenter.previousTrackCommand removeTarget:nil];
    [cmdCenter.changePlaybackPositionCommand removeTarget:nil];
}
*/
import "C"
import (
	"time"
	"unsafe"
)

var globalHandler func(Command, float64)

type darwinController struct{}

func New() Controller {
	return &darwinController{}
}

func (d *darwinController) SetPlaying(title string, duration time.Duration) {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	C.nowplaying_setPlaying(cTitle, C.double(duration.Seconds()))
}

func (d *darwinController) SetStation(name string) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	C.nowplaying_setStation(cName)
}

func (d *darwinController) SetPaused() {
	C.nowplaying_setPaused()
}

func (d *darwinController) SetStopped() {
	C.nowplaying_setStopped()
}

func (d *darwinController) SetProgress(position time.Duration) {
	C.nowplaying_setProgress(C.double(position.Seconds()))
}

func (d *darwinController) SetCommandHandler(handler func(Command, float64)) {
	globalHandler = handler
	if handler != nil {
		C.nowplaying_setCommandHandler()
	} else {
		C.nowplaying_clearCommandHandler()
	}
}

//export goNowPlayingCommand
func goNowPlayingCommand(cmd C.int, value C.double) {
	if globalHandler != nil {
		globalHandler(Command(cmd), float64(value))
	}
}
