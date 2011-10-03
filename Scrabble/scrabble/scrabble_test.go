package scrabble_test

import ("os"; "path/filepath"; "testing";
        "moves"; "scrabble"; "sort_with"; "util")

func TestBlankScore(t *testing.T) {
  if scrabble.BlankScore(10, 5, '-') != 5 {
    t.Fail()
  }
  if scrabble.BlankScore(10, 3, '1') != 4 {
    t.Fail()
  }
  if scrabble.BlankScore(10, 2, '3') != 3 {
    t.Fail()
  }
}

func TestCanFollow(t *testing.T) {
  dict := util.TestInsertIntoDictionary()
  if !scrabble.CanFollow(dict, "AB", map[byte] int {'R': 1}) {
    t.Fail()
  } else if scrabble.CanFollow(dict, "AB", map[byte] int {'C': 1}) {
    t.Fail()
  }
}

func TestGetMoveListAcross(t *testing.T) {
  dict := util.TestInsertIntoDictionary()
  board := [][]byte{
      []byte("4---2--3--2---4"),
      []byte("-3---4---4---3-"),
      []byte("--1---3-3---1--"),
      []byte("---4---1---4---"),
      []byte("2---1-3-3-1---2"),
      []byte("-4---4---4---4-"),
      []byte("--3-3-----3-3--"),
      []byte("3--1---A---1--3"),
      []byte("--3-3-----3-3--"),
      []byte("-4---4---4---4-"),
      []byte("2---1-3-3-1---2"),
      []byte("---4---1---4---"),
      []byte("--1---3-3---1--"),
      []byte("-3---4---4---3-"),
      []byte("4---2--3--2---4")}
  tiles := map[byte] int{'A': 1, 'B': 2, 'R': 1}
  letterValues := map[byte] int{'A': 1, 'B': 1, 'R': 2}
  crossChecks := make(map[int] map[byte] int)

  comparedMoves := []moves.Move {
    moves.Move{"ABRA", 5, moves.Location{7, 4}, moves.ACROSS},
    moves.Move{"ABRA", 5, moves.Location{7, 7}, moves.ACROSS},
    moves.Move{"ABBA", 4, moves.Location{7, 4}, moves.ACROSS},
    moves.Move{"ABBA", 4, moves.Location{7, 7}, moves.ACROSS}}

  moveList := scrabble.GetMoveListAcross(
    dict, board, tiles, letterValues, crossChecks)
  sort_with.SortWith(*moveList, moves.Greater)
  util.RemoveDuplicates(moveList)
  if moveList.Len() != 4 {
    util.PrintMoveList(moveList, board, 25)
    t.Fatalf("length of move list: %d, should have been: 4", moveList.Len())
  }
  for i := 0; i < moveList.Len(); i++ {
    move := moveList.At(i).(moves.Move)
    if !move.Equals(&comparedMoves[i]) {
      moves.PrintMove(&move)
      moves.PrintMove(&comparedMoves[i])
      t.Fatalf("!move.Equals(&comparedMoves[i])")
    }
  }
}

func numTotalTopMoves(
    t *testing.T, board [][]byte, tilesFlag string, num int, score int,
    numTop int) {
  wordListFile, err := os.Open(filepath.Join(os.Getenv("SRCROOT"), "twl.txt"));
  defer wordListFile.Close();
  if err != nil {
    t.Fatal("Could not open twl.txt successfully.")
  }
  dict := util.ReadWordList(wordListFile)
  tiles := util.ReadTiles(tilesFlag)
  letterValues := util.ReadLetterValues(
      "1 4 4 2 1 4 3 4 1 10 5 1 3 1 1 4 10 1 1 1 2 4 4 8 4 10")
  moveList := scrabble.GetMoveList(dict, board, tiles, letterValues)
  if moveList.Len() != num {
    util.PrintMoveList(moveList, board, 25)
    t.Errorf("length of moveList: %d, should have been: %d", moveList.Len(),
             num)
  }
  topMove := moveList.At(0).(moves.Move)
  topMoveScore := topMove.Score
  numTopMoves := 1
  for i := 1; i < moveList.Len(); i++ {
    if moveList.At(i).(moves.Move).Score == topMoveScore {
      numTopMoves++
    } else {
      break
    }
  }
  if topMoveScore != score {
    moves.PrintMove(&topMove)
    t.Errorf("top move score: %d, should have been: %d", topMoveScore, score)
  } else if numTopMoves != numTop {
    t.Errorf("number of top moves: %d, should have been: %d", numTopMoves,
             numTop)
  }
}

func TestNumTotalTopMoves(t *testing.T) {
  board := [][]byte{
    []byte("4---2--3--2---4"),
    []byte("-3---4---4---3-"),
    []byte("--1---3-3---1--"),
    []byte("---4---1---4---"),
    []byte("2---1-3-3-1---2"),
    []byte("-4---4---4---4-"),
    []byte("--3-3-----3-3--"),
    []byte("3--1---*---1--3"),
    []byte("--3-3-----3-3--"),
    []byte("-4---4---4---4-"),
    []byte("2---1-3-3-1---2"),
    []byte("---4---1---4---"),
    []byte("--1---3-3---1--"),
    []byte("-3---4---4---3-"),
    []byte("4---2--3--2---4")}
  numTotalTopMoves(t, board, "ABCDEFG", 346, 24, 8)
  numTotalTopMoves(t, board, "ABCDEF ", 4816, 28, 8)
  board[7] = []byte("3--1-FACED-1--3")
  numTotalTopMoves(t, board, "ABCDEFG", 337, 34, 1)
  board[8][7] = byte('A')
  board[9][7] = byte('R')
  numTotalTopMoves(t, board, "ABCDEF ", 5355, 45, 1)
}

