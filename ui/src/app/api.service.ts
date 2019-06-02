import { Injectable } from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {retry} from 'rxjs/operators';
import {Observable} from 'rxjs';

export interface CreateRequest {
  url: string;
  format: string;
  quality: string;
  page_size: number;
}

export interface CreateResponse {
  id: string;
}

export interface UserResponse {
  user_id: string;
  feature_level: number;
  full_name: string;
}

@Injectable({
  providedIn: 'root'
})
export class APIService {
  constructor(private http: HttpClient) {}

  createFeed(request: CreateRequest): Observable<CreateResponse> {
    return this.http.post<CreateResponse>('/api/create', request);
  }

  getUser(): Observable<UserResponse> {
    return this.http.get<UserResponse>('/api/user')
      .pipe(
        retry(3)
      );
  }
}
