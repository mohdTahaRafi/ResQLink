import 'dart:convert';
import 'package:dio/dio.dart';
import 'package:firebase_auth/firebase_auth.dart';

/// HTTP client for the SAMAJ Go backend API.
class ApiClient {
  late final Dio _dio;

  ApiClient({required String baseUrl}) {
    _dio = Dio(BaseOptions(
      baseUrl: baseUrl,
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 15),
      headers: {'Content-Type': 'application/json'},
    ));

    _dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) async {
        final user = FirebaseAuth.instance.currentUser;
        if (user != null) {
          final token = await user.getIdToken();
          options.headers['Authorization'] = 'Bearer $token';
        }
        handler.next(options);
      },
      onError: (error, handler) {
        if (error.response?.statusCode == 401) {
          FirebaseAuth.instance.signOut();
        }
        handler.next(error);
      },
    ));
  }

  /// Submit a new field report.
  Future<Map<String, dynamic>> createReport({
    required String rawText,
    String? mediaUrl,
    String? mediaType,
    required double latitude,
    required double longitude,
  }) async {
    final response = await _dio.post('/api/v1/reports', data: {
      'raw_text': rawText,
      'media_url': mediaUrl ?? '',
      'media_type': mediaType ?? 'text',
      'latitude': latitude,
      'longitude': longitude,
    });
    return response.data;
  }

  /// Fetch reports list.
  Future<List<dynamic>> getReports({int limit = 50}) async {
    final response = await _dio.get('/api/v1/reports', queryParameters: {'limit': limit});
    return response.data['reports'] ?? [];
  }

  /// Fetch role-specific dashboard data.
  Future<Map<String, dynamic>> getDashboard(String role) async {
    final response = await _dio.get('/api/v1/dashboard/$role');
    return response.data;
  }

  /// Register as a volunteer.
  Future<Map<String, dynamic>> createVolunteer({
    required String name,
    required List<String> skills,
    required double latitude,
    required double longitude,
  }) async {
    final response = await _dio.post('/api/v1/volunteers', data: {
      'name': name,
      'skills': skills,
      'latitude': latitude,
      'longitude': longitude,
      'available': true,
    });
    return response.data;
  }

  /// Get volunteer matches for a ward.
  Future<Map<String, dynamic>> getMatches(String wardId, {List<String>? skills}) async {
    final params = <String, dynamic>{};
    if (skills != null && skills.isNotEmpty) {
      params['skills'] = jsonEncode(skills);
    }
    final response = await _dio.get('/api/v1/match/$wardId', queryParameters: params);
    return response.data;
  }
}
