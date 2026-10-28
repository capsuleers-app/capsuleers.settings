//go:build linux && gtk3

package main

/*
#cgo pkg-config: gtk+-3.0
#include <gtk/gtk.h>
#include <gdk/gdkwayland.h>

// capsFramelessWayland makes a GTK3 window truly frameless on Wayland.
//
// gtk_window_set_decorated(FALSE) (what Wails does for Frameless) tells GTK3 not to draw a
// titlebar but, on Wayland, GTK3 then never announces that it draws its own decorations: KWin
// (and other compositors with server side decorations) adds its own title bar on top of the
// custom one. A window with a titlebar widget is announced as client side decorated, so the
// compositor adds nothing. The titlebar is an empty, hidden box and the CSD shadow, border
// and rounded corners are removed so only the web content is visible.
// Returns 1 when applied (Wayland display only; X11 already honours decorated=FALSE).
static int capsFramelessWayland(void *p) {
	GtkWindow *w = GTK_WINDOW(p);
	if (!w || !GDK_IS_WAYLAND_DISPLAY(gtk_widget_get_display(GTK_WIDGET(w)))) return 0;
	if (gtk_widget_get_realized(GTK_WIDGET(w))) return 0;
	GtkWidget *bar = gtk_box_new(GTK_ORIENTATION_HORIZONTAL, 0);
	gtk_widget_set_no_show_all(bar, TRUE);
	gtk_widget_set_visible(bar, FALSE);
	gtk_window_set_titlebar(w, bar);
	gtk_window_set_decorated(w, TRUE);
	GtkCssProvider *css = gtk_css_provider_new();
	gtk_css_provider_load_from_data(css,
		"window.capsuleers-frameless, window.capsuleers-frameless decoration {"
		" box-shadow: none; margin: 0; border: none; border-radius: 0; }", -1, NULL);
	GtkStyleContext *ctx = gtk_widget_get_style_context(GTK_WIDGET(w));
	gtk_style_context_add_class(ctx, "capsuleers-frameless");
	gtk_style_context_add_provider_for_screen(gtk_widget_get_screen(GTK_WIDGET(w)),
		GTK_STYLE_PROVIDER(css), GTK_STYLE_PROVIDER_PRIORITY_USER);
	g_object_unref(css);
	return 1;
}
*/
import "C"

import (
	"sync"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

// fixWaylandFrame runs once, while the window is still hidden (before GTK realizes it).
func init() { framelessFix = fixWaylandFrame }

func fixWaylandFrame(win *application.WebviewWindow) {
	var once sync.Once
	reveal := func() {
		once.Do(func() {
			application.InvokeSync(func() {
				if p := win.NativeWindow(); p != nil {
					C.capsFramelessWayland(unsafe.Pointer(p))
				}
			})
			win.Show()
		})
	}
	win.RegisterHook(events.Common.WindowRuntimeReady, func(*application.WindowEvent) { reveal() })
	// never leave the window hidden if the runtime handshake does not arrive
	time.AfterFunc(5*time.Second, reveal)
}
