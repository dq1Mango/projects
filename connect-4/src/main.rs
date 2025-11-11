use colored_text::Colorize;
use std::io;

#[derive(Default, Clone, PartialEq)]
enum Player {
  #[default]
  Player1,
  Player2,
}

struct Collum {
  pieces: Vec<Option<Player>>,
  quantity: usize,
  capacity: usize,
}

impl Collum {
  fn new(length: usize) -> Collum {
    let mut pieces = Vec::new();
    for _ in 0..length {
      pieces.push(None);
    }

    Self {
      pieces: pieces,
      quantity: 0,
      capacity: length,
    }
  }

  fn can_play(&self) -> bool {
    self.quantity < self.capacity
  }

  fn drop(&mut self, piece: Player) {
    if !self.can_play() {
      panic!("Aaaahhhh");
    }

    self.pieces[self.quantity] = Some(piece);
    self.quantity += 1;
  }
}

#[derive(Default)]
struct State {
  board: Vec<Collum>,
  rows: usize,
  cols: usize,
  turn: Player,
}

impl State {
  fn new(rows: usize, cols: usize) -> Self {
    let mut state = Self::default();

    for _ in 0..cols {
      state.board.push(Collum::new(rows));
    }

    state.rows = rows;
    state.cols = cols;

    state
  }

  fn advance_turn(&mut self) {
    if self.turn == Player::Player1 {
      self.turn = Player::Player2;
    } else {
      self.turn = Player::Player1;
    }
  }

  fn detect_win_for_player(&self, player: Player) -> bool {
    let player_option = Some(player);
    let mut in_a_row;

    // row win
    for row in 0..self.rows {
      for col in 0..self.cols - 4 {
        in_a_row = true;
        for i in 0..4 {
          if self.board[col + i].pieces[row] != player_option {
            in_a_row = false;
            break;
          }
        }

        if in_a_row {
          return true;
        }
      }
    }

    // collumn win
    for col in 0..self.cols {
      for row in 0..self.rows - 4 {
        in_a_row = true;
        for i in 0..4 {
          if self.board[col].pieces[row + i] != player_option {
            in_a_row = false;
            break;
          }
        }
        if in_a_row {
          return true;
        }
      }
    }

    // diagonal up-right / left-down win
    for col in 0..self.cols - 4 {
      for row in 0..self.rows - 4 {
        in_a_row = true;
        for i in 0..4 {
          if self.board[col + i].pieces[row + i] != player_option {
            in_a_row = false;
            break;
          }
        }
        if in_a_row {
          return true;
        }
      }
    }

    // diagonal up-left / down - right win
    for col in 0..self.cols - 4 {
      for row in 3..self.rows {
        in_a_row = true;
        for i in 0..4 {
          if self.board[col + i].pieces[row - i] != player_option {
            in_a_row = false;
            break;
          }
        }
        if in_a_row {
          return true;
        }
      }
    }

    false
  }
  fn detect_win(&self) -> Option<Player> {
    if self.detect_win_for_player(Player::Player1) {
      return Some(Player::Player1);
    } else if self.detect_win_for_player(Player::Player2) {
      return Some(Player::Player2);
    } else {
      return None;
    }
  }

  fn make_move(&mut self, position: usize) -> Option<Player> {
    if !self.board[position].can_play() {
      panic!("aahhhhh");
    }

    self.board[position].drop(self.turn.clone());

    if self.detect_win_for_player(self.turn.clone()) {
      return Some(self.turn.clone());
    }

    self.advance_turn();

    None
  }

  fn display(&self) {
    let background_color = "#45475a";
    // let piece = "o";
    let piece = "";

    print!(" ");
    for i in 0..self.cols {
      print!("{}", i + 1);
      print!(" ");
    }

    println!();

    for row in (0..self.rows).rev() {
      for col in 0..self.cols {
        print!("{}", "|".on_hex(background_color));
        match self.board[col].pieces[row] {
          Some(Player::Player1) => {
            print!("{}", piece.hex("#89b4fa").on_hex(background_color));
          }
          Some(Player::Player2) => {
            print!("{}", piece.hex("#94e2d5").on_hex(background_color));
          }
          None => {
            print!("{}", " ".on_hex(background_color));
          }
        }
      }
      println!("{}", "|".on_hex(background_color));
    }
  }
}
fn clear_screen() {
  print!("\x1b[2J");
  print!("\x1b[2H");
}

fn main() -> io::Result<()> {
  let cols = 7;
  let rows = 6;

  let mut state = State::new(rows, cols);
  let mut buffer;
  let stdin = io::stdin(); // We get `Stdin` here.
  loop {
    clear_screen();
    state.display();

    let player_num = match state.turn {
      Player::Player1 => 1,
      Player::Player2 => 2,
    };

    let mut choice: usize;

    loop {
      println!("Player #{} make your move: ", player_num);
      buffer = String::new();
      stdin.read_line(&mut buffer)?;
      match buffer.trim().parse::<usize>() {
        Ok(x) => {
          choice = x.clamp(1, 7) - 1;
          if state.board[choice].can_play() {
            break;
          } else {
            println!("Invalid move!");
          }
        }
        Err(_) => println!("Must be a number between 1 and 7"),
      }
    }

    let win = state.make_move(choice);
    if let Some(_) = win {
      clear_screen();
      state.display();
      println!("Contadulation player __, you win!!!");
      break;
    }
  }
  Ok(())
}
