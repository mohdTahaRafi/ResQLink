// File generated for Project RESQLINK (resqlink-58742).
// Values extracted from google-services.json.
import 'package:firebase_core/firebase_core.dart' show FirebaseOptions;
import 'package:flutter/foundation.dart'
    show defaultTargetPlatform, kIsWeb, TargetPlatform;

class DefaultFirebaseOptions {
  static FirebaseOptions get currentPlatform {
    if (kIsWeb) {
      return web;
    }
    switch (defaultTargetPlatform) {
      case TargetPlatform.android:
        return android;
      case TargetPlatform.iOS:
        return ios;
      case TargetPlatform.macOS:
        return macos;
      case TargetPlatform.windows:
        return android; // fallback
      case TargetPlatform.linux:
        return android; // fallback
      case TargetPlatform.fuchsia:
        throw UnsupportedError('Fuchsia is not supported');
    }
  }

  static const FirebaseOptions android = FirebaseOptions(
    apiKey: 'AIzaSyCEGxHt5gmtlCTa1AJo54WILEt7iYlTMTc',
    appId: '1:668252030235:android:80e828250f300ae564bfbe',
    messagingSenderId: '668252030235',
    projectId: 'resqlink-58742',
    storageBucket: 'resqlink-58742.firebasestorage.app',
  );

  // TODO: Add iOS app in Firebase Console and fill these in
  static const FirebaseOptions ios = FirebaseOptions(
    apiKey: 'AIzaSyCEGxHt5gmtlCTa1AJo54WILEt7iYlTMTc',
    appId: '1:668252030235:android:80e828250f300ae564bfbe',
    messagingSenderId: '668252030235',
    projectId: 'resqlink-58742',
    storageBucket: 'resqlink-58742.firebasestorage.app',
    iosBundleId: 'com.resqlink.mobile',
  );

  // Web app registered in Firebase Console
  static const FirebaseOptions web = FirebaseOptions(
    apiKey: 'AIzaSyAGGoqlpORmpAA8I5hQs4c5cBFf31Ipv5A',
    appId: '1:668252030235:web:f65713fc301d062864bfbe',
    messagingSenderId: '668252030235',
    projectId: 'resqlink-58742',
    storageBucket: 'resqlink-58742.firebasestorage.app',
    authDomain: 'resqlink-58742.firebaseapp.com',
    measurementId: 'G-DWZK4V0PYZ',
  );

  static const FirebaseOptions macos = FirebaseOptions(
    apiKey: 'AIzaSyCEGxHt5gmtlCTa1AJo54WILEt7iYlTMTc',
    appId: '1:668252030235:android:80e828250f300ae564bfbe',
    messagingSenderId: '668252030235',
    projectId: 'resqlink-58742',
    storageBucket: 'resqlink-58742.firebasestorage.app',
    iosBundleId: 'com.resqlink.mobile',
  );
}
