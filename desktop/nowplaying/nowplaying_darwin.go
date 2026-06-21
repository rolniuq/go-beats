//go:build desktop

package nowplaying

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework MediaPlayer -framework Cocoa
#include <stdlib.h>
#import <Cocoa/Cocoa.h>
#import <MediaPlayer/MediaPlayer.h>

void nowplaying_setPlaying(const char *title, double duration) {
	@autoreleasepool {
		MPNowPlayingInfoCenter *center = MPNowPlayingInfoCenter.defaultCenter;
		NSString *t = [NSString stringWithUTF8String:title];
		NSImage *icon = [NSImage imageNamed:@"AppIcon"] ?: [NSImage imageNamed:@"NSApplicationIcon"];
		MPMediaItemArtwork *art = nil;
		if (icon) {
			art = [[MPMediaItemArtwork alloc] initWithBoundsSize:icon.size requestHandler:^NSImage*(CGSize s){ return icon; }];
		}
		NSMutableDictionary *info = [NSMutableDictionary dictionary];
		info[MPMediaItemPropertyTitle] = t;
		info[MPMediaItemPropertyArtist] = @"go-beats";
		info[MPMediaItemPropertyAlbumTitle] = @"Lofi Beats";
		info[MPNowPlayingInfoPropertyPlaybackRate] = @(1.0);
		if (duration > 0) info[MPMediaItemPropertyPlaybackDuration] = @(duration);
		if (art) info[MPMediaItemPropertyArtwork] = art;
		center.nowPlayingInfo = info;
	}
}

void nowplaying_setStation(const char *name) {
	@autoreleasepool {
		MPNowPlayingInfoCenter *center = MPNowPlayingInfoCenter.defaultCenter;
		NSString *n = [NSString stringWithUTF8String:name];
		NSImage *icon = [NSImage imageNamed:@"AppIcon"] ?: [NSImage imageNamed:@"NSApplicationIcon"];
		MPMediaItemArtwork *art = nil;
		if (icon) {
			art = [[MPMediaItemArtwork alloc] initWithBoundsSize:icon.size requestHandler:^NSImage*(CGSize s){ return icon; }];
		}
		NSMutableDictionary *info = [NSMutableDictionary dictionary];
		info[MPMediaItemPropertyTitle] = n;
		info[MPMediaItemPropertyArtist] = @"go-beats";
		info[MPMediaItemPropertyAlbumTitle] = @"Radio";
		info[MPNowPlayingInfoPropertyPlaybackRate] = @(1.0);
		info[MPMediaItemPropertyPlaybackDuration] = @(-1);
		if (art) info[MPMediaItemPropertyArtwork] = art;
		center.nowPlayingInfo = info;
	}
}

void nowplaying_setPaused(void) {
	@autoreleasepool {
		MPNowPlayingInfoCenter *center = MPNowPlayingInfoCenter.defaultCenter;
		NSMutableDictionary *info = [center.nowPlayingInfo mutableCopy];
		if (info) {
			info[MPNowPlayingInfoPropertyPlaybackRate] = @(0.0);
			center.nowPlayingInfo = info;
		}
	}
}

void nowplaying_setStopped(void) {
	MPNowPlayingInfoCenter.defaultCenter.nowPlayingInfo = nil;
}

void nowplaying_setProgress(double position) {
	@autoreleasepool {
		MPNowPlayingInfoCenter *center = MPNowPlayingInfoCenter.defaultCenter;
		NSMutableDictionary *info = [center.nowPlayingInfo mutableCopy];
		if (info) {
			info[MPNowPlayingInfoPropertyElapsedPlaybackTime] = @(position);
			center.nowPlayingInfo = info;
		}
	}
}

extern void goNowPlayingCommand(int cmd, double val);

void nowplaying_setCommandHandler(void) {
	MPRemoteCommandCenter *cc = MPRemoteCommandCenter.sharedCommandCenter;
	[cc.playCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *_Nonnull e){
		goNowPlayingCommand(0, 0); return MPRemoteCommandHandlerStatusSuccess;
	}];
	[cc.pauseCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *_Nonnull e){
		goNowPlayingCommand(1, 0); return MPRemoteCommandHandlerStatusSuccess;
	}];
	[cc.togglePlayPauseCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *_Nonnull e){
		goNowPlayingCommand(2, 0); return MPRemoteCommandHandlerStatusSuccess;
	}];
	[cc.nextTrackCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *_Nonnull e){
		goNowPlayingCommand(3, 0); return MPRemoteCommandHandlerStatusSuccess;
	}];
	[cc.previousTrackCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *_Nonnull e){
		goNowPlayingCommand(4, 0); return MPRemoteCommandHandlerStatusSuccess;
	}];
	[cc.changePlaybackPositionCommand addTargetWithHandler:^MPRemoteCommandHandlerStatus(MPRemoteCommandEvent *_Nonnull e){
		goNowPlayingCommand(5, ((MPChangePlaybackPositionCommandEvent *)e).positionTime);
		return MPRemoteCommandHandlerStatusSuccess;
	}];
}

void nowplaying_clearCommandHandler(void) {
	MPRemoteCommandCenter *cc = MPRemoteCommandCenter.sharedCommandCenter;
	[cc.playCommand removeTarget:nil];
	[cc.pauseCommand removeTarget:nil];
	[cc.togglePlayPauseCommand removeTarget:nil];
	[cc.nextTrackCommand removeTarget:nil];
	[cc.previousTrackCommand removeTarget:nil];
	[cc.changePlaybackPositionCommand removeTarget:nil];
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
