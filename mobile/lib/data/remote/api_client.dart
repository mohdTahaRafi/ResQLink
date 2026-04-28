import 'package:dio/dio.dart';
import 'package:firebase_auth/firebase_auth.dart';

/// HTTP client for the SAMAJ Go backend API.
class ApiClient {
  late final Dio _dio;

  ApiClient({required String baseUrl}) {
    _dio = Dio(BaseOptions(
      baseUrl: baseUrl,
      connectTimeout: const Duration(seconds: 10),
      receiveTimeout: const Duration(seconds: 30),
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

  // ══════════════════════════════════════════════════
  // REPORTER — Submit Reports
  // ══════════════════════════════════════════════════

  /// Submit a new issue report with all fields.
  Future<Map<String, dynamic>> createReport({
    required String rawText,
    String? mediaUrl,
    String? mediaType,
    required double latitude,
    required double longitude,
    String issueType = 'civic_issue',
    String userUrgency = 'normal',
    int requiredVolunteers = 1,
    String location = '',
  }) async {
    final response = await _dio.post('/api/v1/reports', data: {
      'raw_text': rawText,
      'media_url': mediaUrl ?? '',
      'media_type': mediaType ?? 'text',
      'latitude': latitude,
      'longitude': longitude,
      'issue_type': issueType,
      'user_urgency': userUrgency,
      'required_volunteers': requiredVolunteers,
      'location': location,
    });
    return response.data;
  }

  /// Fetch reports list.
  Future<List<dynamic>> getReports({int limit = 50}) async {
    final response =
        await _dio.get('/api/v1/reports', queryParameters: {'limit': limit});
    return response.data['reports'] ?? [];
  }

  // ══════════════════════════════════════════════════
  // VOLUNTEER — Task Management
  // ══════════════════════════════════════════════════

  /// Get tasks assigned to the current volunteer.
  Future<List<dynamic>> getMyTasks() async {
    final response = await _dio.get('/api/v1/volunteers/me/tasks');
    return response.data['tasks'] ?? [];
  }

  /// Update the status of a report (volunteer action).
  Future<void> updateReportStatus(String reportId, String newStatus) async {
    await _dio.patch('/api/v1/reports/$reportId/status', data: {
      'status': newStatus,
    });
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

  // ══════════════════════════════════════════════════
  // SPECIALIST — Case Files & AI Search
  // ══════════════════════════════════════════════════

  /// Get case files assigned to the current specialist.
  Future<List<dynamic>> getMyCases() async {
    final response = await _dio.get('/api/v1/cases/my');
    return response.data['cases'] ?? [];
  }

  /// Ask an AI question about a case's documents.
  Future<Map<String, dynamic>> askCaseQuestion(
      String caseId, String question) async {
    final response = await _dio.post('/api/v1/cases/$caseId/ask', data: {
      'question': question,
    });
    return response.data;
  }

  /// Upload a document to a case file.
  Future<void> uploadCaseDocument(
      String caseId, String fileName, String content, String fileType) async {
    await _dio.post('/api/v1/cases/$caseId/documents', data: {
      'file_name': fileName,
      'content': content,
      'file_type': fileType,
    });
  }

  /// Search within documents of a case.
  Future<List<dynamic>> searchCaseDocuments(String caseId, String query) async {
    final response = await _dio.post('/api/v1/cases/$caseId/search', data: {
      'query': query,
    });
    return response.data['results'] ?? [];
  }

  // ══════════════════════════════════════════════════
  // NGO ADMIN — Issue Management & Matching
  // ══════════════════════════════════════════════════

  /// Get all reports sorted by priority.
  Future<List<dynamic>> getPrioritizedReports() async {
    final response = await _dio.get('/api/v1/reports/prioritized');
    return response.data['reports'] ?? [];
  }

  /// Get matching volunteers for a specific report.
  Future<List<dynamic>> getMatchingVolunteers(String reportId) async {
    final response = await _dio.get('/api/v1/reports/$reportId/match');
    return response.data['matches'] ?? [];
  }

  /// Assign volunteers to a report.
  Future<void> assignVolunteers(
      String reportId, List<String> volunteerIds) async {
    await _dio.post('/api/v1/reports/$reportId/assign', data: {
      'volunteer_ids': volunteerIds,
    });
  }

  /// Fetch role-specific dashboard data.
  Future<Map<String, dynamic>> getDashboard(String role) async {
    final response = await _dio.get('/api/v1/dashboard/$role');
    return response.data;
  }

  // ══════════════════════════════════════════════════
  // AI-POWERED FEATURES
  // ══════════════════════════════════════════════════

  /// Analyze an image using Gemini Vision.
  Future<Map<String, dynamic>> analyzeImage(String base64Image) async {
    final response = await _dio
        .post('/api/v1/ai/analyze-image', data: {'image': base64Image});
    return response.data['analysis'];
  }

  /// Verify report consistency (image vs text).
  Future<Map<String, dynamic>> verifyReport(
      String base64Image, String text) async {
    final response = await _dio.post('/api/v1/ai/verify-report',
        data: {'image': base64Image, 'text': text});
    return response.data['verification'];
  }

  /// Detect duplicate reports.
  Future<Map<String, dynamic>> detectDuplicates(String text) async {
    final response =
        await _dio.post('/api/v1/ai/detect-duplicates', data: {'text': text});
    return response.data['duplicate_check'];
  }

  /// Generate AI action plan for a report.
  Future<Map<String, dynamic>> getActionPlan(String reportId) async {
    final response = await _dio
        .post('/api/v1/ai/action-plan', data: {'report_id': reportId});
    return response.data['action_plan'];
  }

  /// Analyze sentiment of text.
  Future<Map<String, dynamic>> analyzeSentiment(String text) async {
    final response =
        await _dio.post('/api/v1/ai/sentiment', data: {'text': text});
    return response.data['sentiment'];
  }

  /// Translate text between languages.
  Future<Map<String, dynamic>> translate(String text,
      {String sourceLang = '', String targetLang = 'English'}) async {
    final response = await _dio.post('/api/v1/ai/translate', data: {
      'text': text,
      'source_lang': sourceLang,
      'target_lang': targetLang,
    });
    return response.data['translation'];
  }

  /// Get AI-generated progress report.
  Future<Map<String, dynamic>> getProgressReport() async {
    final response = await _dio.get('/api/v1/ai/progress-report');
    return response.data['progress_report'];
  }

  /// Get AI skill recommendations for a volunteer.
  Future<Map<String, dynamic>> getSkillRecommendations(
      List<String> currentSkills, List<String> completedTaskTypes) async {
    final response = await _dio.post('/api/v1/ai/recommend-skills', data: {
      'current_skills': currentSkills,
      'completed_task_types': completedTaskTypes,
    });
    return response.data['recommendations'];
  }

  /// OCR a document image.
  Future<Map<String, dynamic>> ocrDocument(String base64Image) async {
    final response =
        await _dio.post('/api/v1/ai/ocr', data: {'image': base64Image});
    return response.data['ocr_result'];
  }

  /// Chat with AI assistant.
  Future<String> chatWithAI(String message, {String taskContext = ''}) async {
    final response = await _dio.post('/api/v1/ai/chat', data: {
      'message': message,
      'task_context': taskContext,
    });
    return response.data['response'];
  }

  // ══════════════════════════════════════════════════
  // REVERSE GEOCODING
  // ══════════════════════════════════════════════════

  /// Reverse geocode coordinates to a location name.
  Future<String> reverseGeocode(double latitude, double longitude) async {
    final response = await _dio.post('/api/v1/geocode/reverse', data: {
      'latitude': latitude,
      'longitude': longitude,
    });
    return response.data['location'] as String? ?? '';
  }
}
