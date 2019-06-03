import { Component, OnInit } from '@angular/core';
import {APIService, UserResponse} from '../api.service';

@Component({
  selector: 'app-index',
  templateUrl: './index.component.html',
  styleUrls: ['./index.component.scss']
})
export class IndexComponent implements OnInit {
  loggedIn = false;
  user: UserResponse;

  constructor(private apiService: APIService) { }

  ngOnInit() {
    this.apiService.getUser().subscribe(resp => {
      this.loggedIn = true;
      this.user = resp;
    });
  }
}
