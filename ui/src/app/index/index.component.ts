import { Component, OnInit } from '@angular/core';

@Component({
  selector: 'app-index',
  templateUrl: './index.component.html',
  styleUrls: ['./index.component.scss']
})
export class IndexComponent implements OnInit {
  loggedIn: boolean;
  fullName: string;

  constructor() { }

  ngOnInit() {
    this.loggedIn = true;
    this.fullName = 'Full name here';
  }
}
