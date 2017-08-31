import {
  Component,
  HostBinding,
  HostListener,
  OnInit,
  ViewEncapsulation,
  ViewChild,
  AfterViewInit,
  TemplateRef,
  ElementRef
} from '@angular/core';
import {
  ApiService,
  CommonService,
  ThreadPage,
  Post,
  VotesSummary,
  Thread,
  Alert,
  Popup,
  LoadingService,
  FollowPage,
  FollowPageData
} from '../../providers';
import { ActivatedRoute, Router } from '@angular/router';
import { FormControl, FormGroup, Validators } from '@angular/forms';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { slideInLeftAnimation } from '../../animations/router.animations';
import { flyInOutAnimation } from '../../animations/common.animations';
import 'rxjs/add/operator/filter';

@Component({
  selector: 'app-threadpage',
  templateUrl: 'threadPage.component.html',
  styleUrls: ['threadPage.component.scss'],
  encapsulation: ViewEncapsulation.None,
  animations: [slideInLeftAnimation, flyInOutAnimation],
})

export class ThreadPageComponent implements OnInit {
  @HostBinding('@routeAnimation') routeAnimation = true;
  @HostBinding('style.display') display = 'block';
  @ViewChild('editor') editor: ElementRef;
  @ViewChild('fab') fab: TemplateRef<any>;
  sort = 'esc';
  boardKey = '';
  threadKey = '';
  data: ThreadPage;
  postForm = new FormGroup({
    name: new FormControl('', Validators.required),
    body: new FormControl('', Validators.required),
  });
  userFollow: FollowPageData = {};
  showUserInfoMenu = false;
  userTag = '';
  editorOptions = {
    placeholderText: 'Edit Your Content Here!',
    quickInsertButtons: ['table', 'ul', 'ol', 'hr'],
    toolbarButtons: [
      'bold',
      'italic',
      'underline',
      'strikeThrough',
      'subscript',
      'superscript',
      '|',
      'fontFamily',
      'fontSize',
      'color',
      'inlineStyle',
      'paragraphStyle',
      '|',
      'paragraphFormat',
      'align',
      'formatOL',
      'formatUL',
      'outdent',
      'indent',
      'quote',
      '-',
      'emoticons',
      'insertLink',
      '|',
      'insertHR',
      'selectAll',
      'clearFormatting',
      '|',
      'print',
      'spellChecker',
      'help',
      'html',
      '|',
      'undo',
      'redo'],
    heightMin: 200,
    events: {},
  };

  constructor(private api: ApiService,
    private router: Router,
    private route: ActivatedRoute,
    private modal: NgbModal,
    private common: CommonService,
    private alert: Alert,
    private pop: Popup,
    private loading: LoadingService) {
  }

  ngOnInit() {
    this.route.params.subscribe(res => {
      this.boardKey = res['board_public_key'];
      this.threadKey = res['thread_ref'];
      this.open(this.boardKey, this.threadKey);
    });
    // this.common.fb.display = 'flex';
    // this.common.fb.handle = () => {
    //   this.openReply(this.replyBox);
    // }

    this.pop.open(this.fab);
  }
  showUserMenu(post: Post, ev: Event) {
    ev.stopImmediatePropagation();
    ev.stopPropagation();
    ev.preventDefault();
    if (post.creatorMenu) {
      this.userFollow = {};
      post.creatorMenu = false;
      return;
    }
    post.creatorMenu = true;
    this.userFollow = {};
    if (!post.creator) {
      return;
    }
    const data = new FormData();
    data.append('board_public_key', this.boardKey);
    data.append('user_public_key', post.creator);
    this.api.getFollowPage(data).subscribe((res: FollowPage) => {
      if (res.okay) {
        this.userFollow = res.data;
        this.showUserInfoMenu = true;
      }
    })
  }
  Menu(ev: Event, post: Post) {
    ev.stopImmediatePropagation();
    ev.stopPropagation();
    ev.preventDefault();
    if (!post.voteMenu) {
      post.voteMenu = true;
    } else {
      post.voteMenu = false;
    }
  }
  // upThread(ev: Event) {
  //   ev.stopImmediatePropagation();
  //   ev.stopPropagation();
  //   ev.preventDefault();
  //   this.data.data.thread.votes.up_votes.count += 1;
  //   const data = new FormData();
  //   data.append('mode', '+1');
  //   this.addThreadVote(data);
  // }
  // downThread(ev: Event) {
  //   ev.stopImmediatePropagation();
  //   ev.stopPropagation();
  //   ev.preventDefault();
  //   this.data.data.thread.votes.down_votes.count += 1;
  //   const data = new FormData();
  //   this.addThreadVote(data);
  // }
  public setSort() {
    this.sort = this.sort === 'desc' ? 'asc' : 'desc';
  }
  trackPosts(index, post) {
    return post ? post.ref : undefined;
  }
  addThreadVote(mode: string, ev: Event) {
    ev.stopImmediatePropagation();
    ev.stopPropagation();
    ev.preventDefault();
    if (!this.data.data.thread.votes) {
      this.data.data.thread.votes = {
        up_votes: { count: 0, voted: false },
        down_votes: { count: 0, voted: false }
      }
    }
    if (mode === '-1') {
      this.data.data.thread.votes.down_votes.count += 1;
    } else {
      this.data.data.thread.votes.up_votes.count += 1;
    }
    const data = new FormData();
    data.append('board_public_key', this.boardKey);
    data.append('thread_ref', this.threadKey);
    data.append('mode', mode);
    this.api.addThreadVote(data).subscribe(voteRes => {
      if (voteRes.okay) {
        this.data.data.thread.votes = voteRes.data.votes;
      }
    }, err => {
      if (mode === '-1') {
        this.data.data.thread.votes.down_votes.count -= 1;
      } else {
        this.data.data.thread.votes.up_votes.count -= 1;
      }
    })
  }
  addUserVote(ev: Event, user_public_key, mode: string, tag?: string) {
    ev.stopImmediatePropagation();
    ev.stopPropagation();
    ev.preventDefault();
    if (!tag) {
      if (this.userTag === '') {
        tag = 'great';
      } else {
        tag = this.userTag;
      }
    }
    const data = new FormData();
    data.append('board_public_key', this.boardKey);
    data.append('user_public_key', user_public_key);
    data.append('mode', mode);
    data.append('tag', tag);
    this.api.addUserVote(data).subscribe(result => {
      console.log('vote user:', result);
      if (result.okay) {
        this.userFollow = result.data;
      }
    }, err => {
    })
  }
  addPostVote(mode: string, post: Post, ev: Event) {
    ev.stopImmediatePropagation();
    ev.stopPropagation();
    post.voteMenu = false;
    if (post.uiOptions !== undefined && post.uiOptions.voted !== undefined && post.uiOptions.voted) {
      return;
    }
    if (!post.votes) {
      post.votes = {
        up_votes: { count: 0, voted: false },
        down_votes: { count: 0, voted: false }
      }
    }
    if (mode === '-1') {
      post.votes.down_votes.count += 1;
    } else {
      post.votes.up_votes.count += 1;
    }
    const data = new FormData();
    data.append('board_public_key', this.boardKey);
    data.append('post_ref', post.ref);
    data.append('mode', mode);
    this.api.addPostVote(data).subscribe(res => {
      if (res.okay) {
        post.votes = res.data.votes;
      }
    }, err => {
      if (mode === '-1') {
        post.votes.down_votes.count -= 1;
      } else {
        post.votes.up_votes.count -= 1;
      }
    })
  }
  openReply(content) {
    this.postForm.reset();
    this.modal.open(content, { backdrop: 'static', size: 'lg', keyboard: false }).result.then((result) => {
      if (result) {
        if (!this.postForm.valid) {
          this.alert.error({ content: 'title and content can not be empty' });
          return;
        }
        const data = new FormData();
        data.append('board_public_key', this.boardKey);
        data.append('thread_ref', this.threadKey);
        data.append('name', this.postForm.get('name').value);
        data.append('body', this.postForm.get('body').value);
        this.loading.start();
        this.api.newPost(data).subscribe((res: ThreadPage) => {
          if (res.okay) {
            this.data.data.posts = res.data.posts;
            this.alert.success({ content: 'Added successfully' });
            this.loading.close();
          }
        });
      }
    }, err => {
    });

  }
  PostAuthorMenu(post: Post, ev: Event) {
    ev.stopImmediatePropagation();
    ev.stopPropagation();
    ev.preventDefault();
    if (!post.uiOptions) {
      post.uiOptions = { menu: true };
    } else {
      post.uiOptions.menu = !post.uiOptions.menu;
    }
  }
  open(boardKey, ref: string) {
    if (boardKey === '' || ref === '') {
      this.alert.error({ content: 'Parameter error!!!' });
      return;
    }
    const data = new FormData();
    data.append('board_public_key', boardKey);
    data.append('thread_ref', ref);
    this.api.getThreadpage(data).subscribe(res => {
      this.data = res;
    }, err => {
      this.router.navigate(['']);
    });
  }

}
